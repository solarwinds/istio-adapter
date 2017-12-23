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
	bucketName = "docker"

	defaultRetention = "24h"

	defaultTemplate = `{{or (.originIp) "-"}} - {{or (.sourceUser) "-"}} [{{or (.timestamp.Format "2006-01-02T15:04:05Z07:00") "-"}}] "{{or (.method) "-"}} {{or (.url) "-"}} {{or (.protocol) "-"}}" {{or (.responseCode) "-"}} {{or (.responseSize) "-"}}`
)

var db *bolt.DB

type logInfo struct {
	labels []string
	tmpl   *template.Template
}
type PaperTrailLogger struct {
	paperTrailURL string

	retentionPeriod time.Duration

	writer *syslog.Writer

	mu sync.Mutex

	db *bolt.DB

	logInfos map[string]*logInfo

	log adapter.Logger
}

func NewPaperTrailLogger(paperTrailURL string, logRetentionStr string, logConfigs []*config.Params_LogInfo, logger adapter.Logger) (*PaperTrailLogger, error) {

	retention, err := time.ParseDuration(logRetentionStr)
	if err != nil {
		retention, _ = time.ParseDuration(defaultRetention)
	}
	if retention.Seconds() <= float64(0) {
		retention, _ = time.ParseDuration(defaultRetention)
	}

	// This is being called several times. But BoltDB only allows one connection.
	if db == nil {
		db, err = bolt.Open("istio-papertrail.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			logger.Errorf("Unable to open a database for papertrail log processing: %v.", err)
			return nil, err
		}
	}

	logger.Infof("Creating a new paper trail logger for url: %s", paperTrailURL)

	p := &PaperTrailLogger{
		paperTrailURL:   paperTrailURL,
		retentionPeriod: time.Duration(retention) * time.Hour,
		db:              db,
		log:             logger,
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

func (p *PaperTrailLogger) Log(msg *logentry.Instance) error {
	p.log.Infof("AO - In Log method. Received msg: %s", msg)
	linfo, ok := p.logInfos[msg.Name]
	if !ok {
		err := fmt.Errorf("Got an unknown instance of log: %s. Hence Skipping.", msg.Name)
		p.log.Errorf("%v", err)
		return err
	}
	buf := pool.GetBuffer()
	msg.Variables["timestamp"] = time.Now()
	ipval, ok := msg.Variables["originIp"].([]byte)
	if ok {
		msg.Variables["originIp"] = net.IP(ipval).String()
	}

	if err := linfo.tmpl.Execute(buf, msg.Variables); err != nil {
		// We'll just continue on with an empty payload for this entry - we could still be populating the HTTP req with valuable info, for example.
		p.log.Errorf("failed to execute template for log '%s': %v", msg.Name, err)
	}
	payload := buf.String()
	pool.PutBuffer(buf)

	if len(payload) > 0 {
		p.log.Infof("AO - In Log method. Now persisting to boltdb: %s", msg)
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
			e := fmt.Errorf("Unable to store the log for further processing: %s - Error: %v", payload, err)
			p.log.Errorf("%v", e)
			return e
		}
	}
	return nil
}
func (p *PaperTrailLogger) sendLogs(data []byte) error {
	var err error
	p.log.Infof("AO - In sendLogs method. sending msg: %s", string(data))
	writer, err := syslog.Dial("udp", p.paperTrailURL, syslog.LOG_EMERG|syslog.LOG_KERN, "istio")
	if err != nil {
		e := fmt.Errorf("AO - Failed to dial syslog: %v", err)
		p.log.Errorf("%v", e)
		return e
	}
	err = writer.Info(string(data))
	if err != nil {
		e := fmt.Errorf("failed to send log msg to papertrail: %v", err)
		p.log.Errorf("%v", e)
		return e
	}
	return nil
}

// This should be run in a routine
func (p *PaperTrailLogger) flushLogs() {
	var err error
	for p.db != nil {
		err = p.db.Update(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				err1 := fmt.Errorf("Unable to create bucket error: %v", err)
				p.log.Errorf("%v", err1)
				return err1
			}

			b.ForEach(func(k, v []byte) error {
				err = p.sendLogs(v)
				if err == nil {
					p.log.Infof("AO - flushLogs, delete key: %s", string(k))
					err = b.Delete(k)
					if err != nil {
						p.log.Errorf("Unable to delete from boltdb: %v. Continuing to try", err)
					}
				}

				tsN, _ := strconv.ParseInt(string(k), 10, 64)
				ts := time.Unix(0, tsN)

				if time.Since(ts) > p.retentionPeriod {
					p.log.Infof("AO - flushLogs, delete key: %s bcoz it is past retention period.", string(k))
					err = b.Delete(k)
				}
				return err
			})
			return nil
		})
		if err != nil {
			p.log.Errorf("Error reading the data in the DB and shipping logs: %v", err)
		}
		time.Sleep(time.Second)
	}
}
func (p *PaperTrailLogger) Close() error {
	var err error
	if p.writer != nil {
		err = p.writer.Close()
		if err != nil {
			e := fmt.Errorf("failed to close papertrail logger: %v", err)
			p.log.Errorf("%v", e)
			return e
		}
	}
	return err
}
