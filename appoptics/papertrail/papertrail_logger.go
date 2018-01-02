package papertrail

import (
	"fmt"
	"html/template"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"log/syslog"

	"github.com/boltdb/bolt"
	"istio.io/istio/mixer/adapter/appoptics/config"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/logentry"

	"istio.io/istio/mixer/pkg/pool"
)

const (
	bucketName = "istio"

	dbName = "istio-papertrail.db"

	defaultWorkerCount = 10

	defaultRetention = "24h"

	defaultTemplate = `{{or (.originIp) "-"}} - {{or (.sourceUser) "-"}} [{{or (.timestamp.Format "2006-01-02T15:04:05Z07:00") "-"}}] "{{or (.method) "-"}} {{or (.url) "-"}} {{or (.protocol) "-"}}" {{or (.responseCode) "-"}} {{or (.responseSize) "-"}}`
)

var db *bolt.DB

type logInfo struct {
	labels []string
	tmpl   *template.Template
}

type PaperTrailLoggerInterface interface {
	Log(*logentry.Instance) error
	Close() error
}

const maxRetry = 1

type PaperTrailLogger struct {
	paperTrailURL string

	retentionPeriod time.Duration

	writer *syslog.Writer

	mu sync.Mutex

	db *bolt.DB

	logInfos map[string]*logInfo

	log adapter.Logger

	maxWorkers int

	loopFactor *bool
}

func NewPaperTrailLogger(paperTrailURL string, logRetentionStr string, logConfigs []*config.Params_LogInfo, logger adapter.Logger, loopFactor *bool) (PaperTrailLoggerInterface, error) {

	retention, err := time.ParseDuration(logRetentionStr)
	if err != nil {
		retention, _ = time.ParseDuration(defaultRetention)
	}
	if retention.Seconds() <= float64(0) {
		retention, _ = time.ParseDuration(defaultRetention)
	}

	// This is being called several times. But BoltDB only allows one connection.
	if db == nil {
		db, err = openDB(dbName, logger)
		if err != nil {
			return nil, err
		}
	}

	logger.Infof("Creating a new paper trail logger for url: %s", paperTrailURL)

	p := &PaperTrailLogger{
		paperTrailURL:   paperTrailURL,
		retentionPeriod: time.Duration(retention) * time.Hour,
		db:              db,
		log:             logger,
		maxWorkers:      defaultWorkerCount,
		loopFactor:      loopFactor,
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
	return p, nil
}

func openDB(dbName string, logger adapter.Logger) (*bolt.DB, error) {
	db, err := bolt.Open(dbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logger.Errorf("Unable to open a database for papertrail log processing: %v.", err)
		return nil, err
	}
	return db, nil
}

func (p *PaperTrailLogger) Log(msg *logentry.Instance) error {
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
	}
	payload := buf.String()
	pool.PutBuffer(buf)

	if len(payload) > 0 {
		retryCount := 0
	RETRY_STORE:
		if retryCount <= maxRetry {
			if p.log.VerbosityLevel(config.DebugLevel) {
				p.log.Infof("AO - In Log method. Now persisting to boltdb: %s", msg)
			}
			err := p.db.Update(func(tx *bolt.Tx) error {
				buc, err := tx.CreateBucketIfNotExists([]byte(bucketName))
				if err != nil {
					p.log.Errorf("Unable to create bucket error: %v", err)
					return err
				}
				err = buc.Put([]byte(fmt.Sprintf("%d", time.Now().UnixNano())), []byte(payload))
				return err
			})
			if err != nil {
				if err == bolt.ErrDatabaseNotOpen {
					db, err = openDB(dbName, p.log)
					if err != nil {
						p.log.Errorf("Error reading the data in the DB and shipping logs: %v", err)
					} else {
						p.db = db
						// retry persisting
						retryCount++
						goto RETRY_STORE
					}
				} else {
					return p.log.Errorf("Unable to store the log for further processing: %s - Error: %v", payload, err)
				}
			}
		} else {
			p.log.Errorf("Max retry limit exceeded. Skipping storing log entry to DB: %s", payload)
		}
	}
	return nil
}
func (p *PaperTrailLogger) sendLogs(data []byte) error {
	var err error
	if p.log.VerbosityLevel(config.DebugLevel) {
		p.log.Infof("AO - In sendLogs method. sending msg: %s", string(data))
	}
	writer, err := syslog.Dial("udp", p.paperTrailURL, syslog.LOG_EMERG|syslog.LOG_KERN, "istio")
	if err != nil {
		return p.log.Errorf("AO - Failed to dial syslog: %v", err)
	}
	err = writer.Info(string(data))
	if err != nil {
		return p.log.Errorf("failed to send log msg to papertrail: %v", err)
	}
	return nil
}

// This should be run in a routine
func (p *PaperTrailLogger) flushLogs() {
	var err error
	for p.db != nil && *p.loopFactor {
		err = p.db.Update(func(tx *bolt.Tx) error {
			hose := make(chan []byte, p.maxWorkers)
			// close the channel
			defer close(hose)
			var wg sync.WaitGroup
			// Assume bucket exists and has keys
			b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				return p.log.Errorf("Unable to create bucket error: %v", err)
			}

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
						val := b.Get(key)
						err = p.sendLogs(val)
						if err == nil {
							if p.log.VerbosityLevel(config.DebugLevel) {
								p.log.Infof("AO - flushLogs, delete key: %s", string(key))
							}
							err = b.Delete(key)
							if err != nil {
								p.log.Errorf("Unable to delete key(%s) from boltdb: %v. Continuing to try", string(key), err)
							} else {
								wg.Done()
								continue
							}
						}

						tsN, _ := strconv.ParseInt(string(key), 10, 64)
						ts := time.Unix(0, tsN)

						if time.Since(ts) > p.retentionPeriod {
							if p.log.VerbosityLevel(config.DebugLevel) {
								p.log.Infof("AO - flushLogs, delete key: %s bcoz it is past retention period.", string(key))
							}
							err = b.Delete(key)
							if err != nil {
								p.log.Errorf("Unable to delete from boltdb: %v. Continuing to try", err)
							}
						}
						wg.Done()
					}
				}(i)
			}

			c := b.Cursor()

			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				wg.Add(1)
				hose <- k
			}

			// need to wait for the tasks to complete
			wg.Wait()
			return nil
		})
		if err != nil {
			if err == bolt.ErrDatabaseNotOpen {
				db, err = openDB(dbName, p.log)
				if err != nil {
					p.log.Errorf("Error reading the data in the DB and shipping logs: %v", err)
				} else {
					p.db = db
				}
			} else {
				p.log.Errorf("Error reading the data in the DB and shipping logs: %v", err)
			}
		}
		time.Sleep(time.Second)
	}
}
func (p *PaperTrailLogger) Close() error {
	var err error
	if p.writer != nil {
		err = p.writer.Close()
		if err != nil {
			return p.log.Errorf("failed to close papertrail logger: %v", err)
		}
	}

	if p.db != nil {
		err = p.db.Close()
		p.db = nil
	}
	return err
}
