// Copyright 2017 Istio Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package papertrail

import (
	"fmt"
	"html/template"
	"log/syslog"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dgraph-io/badger"
	"istio.io/istio/mixer/adapter/solarwinds/config"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/logentry"

	"github.com/satori/go.uuid"
	"istio.io/istio/mixer/pkg/pool"
)

const (
	defaultRetention = "24h"

	defaultMaxDiskUsage = 5 // disk usage in percentage

	defaultUltimateMaxDiskUsage = 99 // usage cannot go beyond this percentage value

	defaultBatchSize = 1000 // records

	dbLocation = "./badger"

	cleanUpInterval = 5 * time.Second

	defaultTemplate = `{{or (.originIp) "-"}} - {{or (.sourceUser) "-"}} ` +
		`[{{or (.timestamp.Format "2006-01-02T15:04:05Z07:00") "-"}}] "{{or (.method) "-"}} {{or (.url) "-"}} ` +
		`{{or (.protocol) "-"}}" {{or (.responseCode) "-"}} {{or (.responseSize) "-"}}`
)

var defaultWorkerCount = 10

type logInfo struct {
	labels []string
	tmpl   *template.Template
}

// LoggerInterface is the interface for all Papertrail logger types
type LoggerInterface interface {
	Log(*logentry.Instance) error
	Close() error
}

const (
	keyFormat  = "TS:%d-BODY:%s"
	keyPattern = "TS:(\\d+)-BODY:(.*)"
)

// Logger is a concrete type of LoggerInterface which collects and ships logs to Papertrail
type Logger struct {
	paperTrailURL string

	retentionPeriod time.Duration

	cmap *sync.Map

	db *badger.DB

	logInfos map[string]*logInfo

	initialDiskUsage float64

	log adapter.Logger

	maxWorkers int

	loopFactor bool
}

// NewLogger does some ground work and returns an instance of LoggerInterface
func NewLogger(paperTrailURL string, logRetentionStr string, logConfigs []*config.Params_LogInfo,
	logger adapter.Logger) (LoggerInterface, error) {

	retention := parseRetention(logRetentionStr)

	opts := badger.DefaultOptions
	opts.Dir = dbLocation
	opts.ValueDir = dbLocation
	//opts.ValueLogLoadingMode = options.FileIO

	db, err := badger.Open(opts)
	if err != nil {
		return nil, logger.Errorf("Error: %v", err)
	}

	logger.Infof("Creating a new paper trail logger for url: %s", paperTrailURL)

	p := &Logger{
		paperTrailURL:    paperTrailURL,
		retentionPeriod:  time.Duration(retention) * time.Hour,
		db:               db,
		initialDiskUsage: diskUsage(),
		log:              logger,
		maxWorkers:       defaultWorkerCount * runtime.NumCPU(),
		loopFactor:       true,
	}

	p.logInfos = map[string]*logInfo{}

	for _, l := range logConfigs {
		var templ string
		if strings.TrimSpace(l.PayloadTemplate) != "" {
			templ = l.PayloadTemplate
		} else {
			templ = defaultTemplate
		}
		tmpl, err := template.New(l.InstanceName).Parse(templ)
		if err != nil {
			logger.Errorf("AO - failed to evaluate template for log instance: %s, skipping: %v", l.InstanceName, err)
			continue
		}
		p.logInfos[l.InstanceName] = &logInfo{
			labels: l.LabelNames,
			tmpl:   tmpl,
		}
	}

	go p.flushLogs()
	go p.deleteExcess()
	go p.cleanup()
	return p, nil
}

// Log method receives log messages
func (p *Logger) Log(msg *logentry.Instance) error {
	if p.log.VerbosityLevel(config.DebugLevel) {
		p.log.Infof("AO - In Log method. Received msg: %v", msg)
	}
	linfo, ok := p.logInfos[msg.Name]
	if !ok {
		return p.log.Errorf("Got an unknown instance of log: %s. Hence Skipping.", msg.Name)
	}
	buf := pool.GetBuffer()
	msg.Variables["timestamp"] = time.Now()
	ipval, ok := msg.Variables["originIp"].([]byte)
	if ok {
		msg.Variables["originIp"] = net.IP(ipval).String()
	}

	if err := linfo.tmpl.Execute(buf, msg.Variables); err != nil {
		p.log.Errorf("failed to execute template for log '%s': %v", msg.Name, err)
		// proceeding anyways
	}
	payload := buf.String()
	pool.PutBuffer(buf)

	if len(payload) > 0 {
		if p.log.VerbosityLevel(config.DebugLevel) {
			p.log.Infof("AO - In Log method. Now persisting log: %s", msg)
		}
		uuid := uuid.NewV4()
		err := p.db.Update(func(txn *badger.Txn) error {
			err := txn.SetWithTTL([]byte(fmt.Sprintf(keyFormat, time.Now().UnixNano(), uuid)), []byte(payload), p.retentionPeriod)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return p.log.Errorf("Error: ", err)
		}
	}
	return nil
}

func (p *Logger) sendLogs(data string) error {
	if p.log.VerbosityLevel(config.DebugLevel) {
		p.log.Infof("AO - In sendLogs method. sending msg: %s", string(data))
	}

	writer, err := syslog.Dial("udp", p.paperTrailURL, syslog.LOG_EMERG|syslog.LOG_KERN, "istio")
	if err != nil {
		return p.log.Errorf("AO - Failed to dial syslog: %v", err)
	}
	defer writer.Close()
	err = writer.Info(data)
	if err != nil {
		return p.log.Errorf("failed to send log msg to papertrail: %v", err)
	}
	return nil
}

// This should be run in a routine
func (p *Logger) flushLogs() {
	for p.loopFactor {
		hose := make(chan []byte, p.maxWorkers)
		var wg sync.WaitGroup

		// workers
		for i := 0; i < p.maxWorkers; i++ {
			go func(worker int) {
				if p.log.VerbosityLevel(config.DebugLevel) {
					p.log.Infof("AO - flushlogs, worker %d initialized.", (worker + 1))
					defer p.log.Infof("AO - flushlogs, worker %d signing off.", (worker + 1))
				}

				for key := range hose {
					if p.log.VerbosityLevel(config.DebugLevel) {
						p.log.Infof("AO - flushlogs, worker %d took the job.", (worker + 1))
					}

					err := p.db.Update(func(txn *badger.Txn) error {
						item, err := txn.Get(key)
						if err != nil {
							if err == badger.ErrKeyNotFound {
								return nil
							} else {
								return err
							}
						}
						var val []byte
						val, err = item.ValueCopy(val)
						if err != nil {
							if err == badger.ErrKeyNotFound {
								return nil
							} else {
								return err
							}
						}
						err = p.sendLogs(string(val))
						if err == nil {
							if p.log.VerbosityLevel(config.DebugLevel) {
								p.log.Infof("AO - flushLogs, delete key: %s", key)
							}
							err := txn.Delete(key)
							if err != nil {
								return err
							}
							return nil
						}
						return nil
					})
					if err != nil {
						p.log.Errorf("Error while deleting key: %s - error: %v", key, err)
					}
					wg.Done()
				}
			}(i)
		}

		err := p.db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				k := make([]byte, len(item.Key()))
				copy(k, item.Key())
				wg.Add(1)
				hose <- k
			}
			return nil
		})
		if err != nil {
			p.log.Errorf("AO - flush logs - Error reading keys from db: ", err)
		}
		wg.Wait()
		close(hose)
		time.Sleep(50 * time.Millisecond)
	}
}

func (p *Logger) deleteExcess() {
	for p.loopFactor {
		currentUsage := diskUsage()
		if p.log.VerbosityLevel(config.DebugLevel) {
			p.log.Infof("Current disk usage: %.2f %%", currentUsage)
			p.log.Infof("DB folder size: %.2f MB", computeDirectorySizeInMegs(dbLocation))
		}
		if currentUsage > p.initialDiskUsage+defaultMaxDiskUsage || currentUsage > defaultUltimateMaxDiskUsage {
			// delete from beginning
			iterations := defaultBatchSize
			err := p.db.View(func(txn *badger.Txn) error {
				opts := badger.DefaultIteratorOptions
				opts.PrefetchValues = false
				it := txn.NewIterator(opts)
				defer it.Close()
				for it.Rewind(); it.Valid(); it.Next() {
					item := it.Item()
					k := make([]byte, len(item.Key()))
					copy(k, item.Key())
					txn.Delete(k)
					iterations--
					if iterations < 0 {
						break
					}
				}
				return nil
			})
			if err != nil {
				p.log.Errorf("AO - deleteExcess - Error while deleting - error: %v", err)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// Close - closes the Logger instance
func (p *Logger) Close() error {
	p.loopFactor = false
	time.Sleep(time.Second)
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *Logger) cleanup() {
	for p.loopFactor {
		if p.db != nil {
			if p.log.VerbosityLevel(config.DebugLevel) {
				p.log.Infof("AO - cleanup - running GC")
			}
			p.db.PurgeOlderVersions()
			p.db.RunValueLogGC(0.99)
		}
		time.Sleep(cleanUpInterval)
	}
}

func parseRetention(logRetentionStr string) time.Duration {
	retention, err := time.ParseDuration(logRetentionStr)
	if err != nil {
		retention, _ = time.ParseDuration(defaultRetention)
	}
	if retention.Seconds() <= float64(0) {
		retention, _ = time.ParseDuration(defaultRetention)
	}
	return retention
}

func diskUsage() float64 {
	var stat syscall.Statfs_t
	wd, _ := os.Getwd()
	syscall.Statfs(wd, &stat)
	avail := stat.Bavail * uint64(stat.Bsize)
	used := stat.Blocks * uint64(stat.Bsize)
	return (float64(used) / float64(used+avail)) * 100
}

func computeDirectorySizeInMegs(fullPath string) float64 {
	var sizeAccumulator int64
	filepath.Walk(fullPath, func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			atomic.AddInt64(&sizeAccumulator, file.Size())
		}
		return nil
	})
	return float64(atomic.LoadInt64(&sizeAccumulator)) / (1024 * 1024)
}
