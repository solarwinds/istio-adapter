// Copyright 2018 Istio Authors.
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
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/satori/go.uuid"

	"istio.io/istio/mixer/adapter/solarwinds/config"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/pkg/pool"
	"istio.io/istio/mixer/template/logentry"
)

const (
	keyFormat                   = "TS:%d-BODY:%s"
	defaultMaxDiskUsage         = 5    // disk usage in percentage
	defaultUltimateMaxDiskUsage = 99   // usage cannot go beyond this percentage value
	defaultBatchSize            = 1000 // records
	dbLocation                  = "./badger"
	cleanUpInterval             = 5 * time.Second
	defaultTemplate             = `{{or (.originIp) "-"}} - {{or (.sourceUser) "-"}} ` +
		`[{{or (.timestamp.Format "2006-01-02T15:04:05Z07:00") "-"}}] "{{or (.method) "-"}} {{or (.url) "-"}} ` +
		`{{or (.protocol) "-"}}" {{or (.responseCode) "-"}} {{or (.responseSize) "-"}}`
)

var (
	defaultWorkerCount = 10
	defaultRetention   = 24 * time.Hour
)

type logInfo struct {
	tmpl *template.Template
}

// LoggerInterface is the interface for all Papertrail logger types
type LoggerInterface interface {
	Log(*logentry.Instance) error
	Close() error
}

// Logger is a concrete type of LoggerInterface which collects and ships logs to Papertrail
type Logger struct {
	paperTrailURL string

	retentionPeriod time.Duration

	db *badger.DB

	logInfos map[string]*logInfo

	initialDiskUsage float64

	log adapter.Logger

	env adapter.Env

	maxWorkers int

	loopFactor bool

	loopWait chan struct{}
}

// NewLogger does some ground work and returns an instance of LoggerInterface
func NewLogger(paperTrailURL string, retention time.Duration, logConfigs map[string]*config.Params_LogInfo,
	env adapter.Env) (LoggerInterface, error) {
	logger := env.Logger()
	if retention.Seconds() <= float64(0) {
		retention = defaultRetention
	}
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
		retentionPeriod:  retention,
		log:              logger,
		env:              env,
		maxWorkers:       defaultWorkerCount * runtime.NumCPU(),
		loopFactor:       true,
		db:               db,
		initialDiskUsage: diskUsage(),
		loopWait:         make(chan struct{}),
	}

	p.logInfos = map[string]*logInfo{}

	for inst, l := range logConfigs {
		var templ string
		if strings.TrimSpace(l.PayloadTemplate) != "" {
			templ = l.PayloadTemplate
		} else {
			templ = defaultTemplate
		}
		tmpl, err := template.New(inst).Parse(templ)
		if err != nil {
			_ = logger.Errorf("failed to evaluate template for log instance: %s, skipping: %v", inst, err)
			continue
		}
		p.logInfos[inst] = &logInfo{
			tmpl: tmpl,
		}
	}

	env.ScheduleDaemon(p.flushLogs)
	env.ScheduleDaemon(p.deleteExcess)
	env.ScheduleDaemon(p.cleanup)
	return p, nil
}

// Log method receives log messages
func (p *Logger) Log(msg *logentry.Instance) error {
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
		_ = p.log.Errorf("failed to execute template for log '%s': %v", msg.Name, err)
		// proceeding anyways
	}
	payload := buf.String()
	pool.PutBuffer(buf)

	if len(payload) > 0 {
		// if p.log.VerbosityLevel(config.DebugLevel) {
		// 	p.log.Infof("AO - In Log method. Now persisting log: %s", msg)
		// }
		guuid := uuid.NewV4()
		if err := p.db.Update(func(txn *badger.Txn) error {
			return txn.SetWithTTL([]byte(fmt.Sprintf(keyFormat, time.Now().UnixNano(), guuid)), []byte(payload), p.retentionPeriod)
		}); err != nil {
			return p.log.Errorf("Error persisting log to local db: %v", err)
		}
	}
	return nil
}

func (p *Logger) sendLogs(data string) error {
	writer, err := syslog.Dial("udp", p.paperTrailURL, syslog.LOG_EMERG|syslog.LOG_KERN, "istio")
	if err != nil {
		return p.log.Errorf("Failed to dial syslog: %v", err)
	}
	defer func() { _ = writer.Close() }()
	if err = writer.Info(data); err != nil {
		return p.log.Errorf("failed to send log msg to papertrail: %v", err)
	}
	return nil
}

// This should be run in a routine
func (p *Logger) flushLogs() {
	defer func() {
		p.loopWait <- struct{}{}
	}()
	for p.loopFactor {
		hose := make(chan []byte, p.maxWorkers)
		wg := new(sync.WaitGroup)

		// workers
		for i := 0; i < p.maxWorkers; i++ {
			p.env.ScheduleDaemon(p.flushWorker(hose, wg))
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
			_ = p.log.Errorf("AO - flush logs - Error reading keys from db: %v", err)
		}
		wg.Wait()
		close(hose)
		time.Sleep(50 * time.Millisecond)
	}
}

func (p *Logger) flushWorker(hose chan []byte, wg *sync.WaitGroup) func() {
	return func() {
		// if p.log.VerbosityLevel(config.DebugLevel) {
		// 	p.log.Infof("AO - flushlogs, worker %d initialized.", (worker + 1))
		// 	defer p.log.Infof("AO - flushlogs, worker %d signing off.", (worker + 1))
		// }

		for key := range hose {
			// if p.log.VerbosityLevel(config.DebugLevel) {
			// 	p.log.Infof("AO - flushlogs, worker %d took the job.", (worker + 1))
			// }

			err := p.db.Update(func(txn *badger.Txn) error {
				item, err := txn.Get(key)
				if err != nil {
					if err == badger.ErrKeyNotFound {
						return nil
					}
					return err
				}
				var val []byte
				val, err = item.ValueCopy(val)
				if err != nil {
					if err == badger.ErrKeyNotFound {
						return nil
					}
					return err
				}
				err = p.sendLogs(string(val))
				if err == nil {
					// if p.log.VerbosityLevel(config.DebugLevel) {
					// 	p.log.Infof("AO - flushLogs, delete key: %s", key)
					// }
					err := txn.Delete(key)
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				_ = p.log.Errorf("Error while deleting key: %s - error: %v", key, err)
			}
			wg.Done()
		}
	}
}

func (p *Logger) deleteExcess() {
	for p.loopFactor {
		currentUsage := diskUsage()
		// if p.log.VerbosityLevel(config.DebugLevel) {
		// 	p.log.Infof("Current disk usage: %.2f %%", currentUsage)
		// 	p.log.Infof("DB folder size: %.2f MB", computeDirectorySizeInMegs(dbLocation))
		// }
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
					_ = txn.Delete(k)
					iterations--
					if iterations < 0 {
						break
					}
				}
				return nil
			})
			if err != nil {
				_ = p.log.Errorf("AO - deleteExcess - Error while deleting - error: %v", err)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// Close - closes the Logger instance
func (p *Logger) Close() error {
	p.loopFactor = false
	defer close(p.loopWait)
	time.Sleep(time.Second)
	if p.db != nil {
		return p.db.Close()
	}
	<-p.loopWait
	return nil
}

func (p *Logger) cleanup() {
	for p.loopFactor {
		if p.db != nil {
			// if p.log.VerbosityLevel(config.DebugLevel) {
			// 	p.log.Infof("AO - cleanup - running GC")
			// }
			_ = p.db.PurgeOlderVersions()
			_ = p.db.RunValueLogGC(0.99)
		}
		time.Sleep(cleanUpInterval)
	}
}

func diskUsage() float64 {
	var stat syscall.Statfs_t
	wd, _ := os.Getwd()
	_ = syscall.Statfs(wd, &stat)
	avail := stat.Bavail * uint64(stat.Bsize)
	used := stat.Blocks * uint64(stat.Bsize)
	return (float64(used) / float64(used+avail)) * 100
}

//func computeDirectorySizeInMegs(fullPath string) float64 {
//	var sizeAccumulator int64
//	filepath.Walk(fullPath, func(path string, file os.FileInfo, err error) error {
//		if !file.IsDir() {
//			atomic.AddInt64(&sizeAccumulator, file.Size())
//		}
//		return nil
//	})
//	return float64(atomic.LoadInt64(&sizeAccumulator)) / (1024 * 1024)
//}
