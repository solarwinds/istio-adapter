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

package solarwinds

import (
	"context"
	"html/template"
	"net"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/segmentio/go-loggly"
	papertrail_go "github.com/solarwinds/papertrail-go"
	"istio.io/istio/mixer/adapter/solarwinds/config"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/pkg/pool"
	"istio.io/istio/mixer/template/logentry"
)

const papertTrailDefaultTemplate = `{{or (.originIp) "-"}} - {{or (.sourceUser) "-"}} ` +
	`[{{or (.timestamp.Format "2006-01-02T15:04:05Z07:00") "-"}}] "{{or (.method) "-"}} {{or (.url) "-"}} ` +
	`{{or (.protocol) "-"}}" {{or (.responseCode) "-"}} {{or (.responseSize) "-"}}`

type logHandlerInterface interface {
	handleLogEntry(context.Context, []*logentry.Instance) error
	close() error
}

type logInfo struct {
	tmpl *template.Template
}

type logHandler struct {
	env              adapter.Env
	logger           adapter.Logger
	paperTrailLogger papertrail_go.LoggerInterface
	logglyClient     *loggly.Client
	logInfos         map[string]*logInfo
}

func newLogHandler(ctx context.Context, env adapter.Env, cfg *config.Params) (logHandlerInterface, error) {
	var ppi papertrail_go.LoggerInterface
	var err error
	var lc *loggly.Client
	if strings.TrimSpace(cfg.PapertrailUrl) != "" {
		var retention time.Duration
		if cfg.PapertrailLocalRetentionDuration != nil {
			retention = *cfg.PapertrailLocalRetentionDuration
		}
		if ppi, err = papertrail_go.NewLogger(ctx, cfg.PapertrailProtocol, cfg.PapertrailUrl, "istio", retention); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(cfg.LogglyToken) != "" {
		lc = loggly.New(cfg.LogglyToken)
	}

	logInfos := map[string]*logInfo{}

	for inst, l := range cfg.Logs {
		var templ string
		if strings.TrimSpace(l.PayloadTemplate) != "" {
			templ = l.PayloadTemplate
		} else {
			templ = papertTrailDefaultTemplate
		}
		tmpl, err := template.New(inst).Parse(templ)
		if err != nil {
			_ = env.Logger().Errorf("failed to evaluate template for log instance: %s, skipping: %v", inst, err)
			continue
		}
		logInfos[inst] = &logInfo{
			tmpl: tmpl,
		}
	}

	return &logHandler{
		logger:           env.Logger(),
		env:              env,
		paperTrailLogger: ppi,
		logglyClient:     lc,
		logInfos:         logInfos,
	}, nil
}

func (h *logHandler) handleLogEntry(ctx context.Context, values []*logentry.Instance) error {
	for _, inst := range values {

		// papertrail
		l, _ := h.paperTrailLogger.(*papertrail_go.Logger)
		if l != nil {
			linfo, ok := h.logInfos[inst.Name]
			if !ok {
				return h.env.Logger().Errorf("Got an unknown instance of log: %s. Hence Skipping.", inst.Name)
			}
			buf := pool.GetBuffer()
			inst.Variables["timestamp"] = time.Now()
			ipval, ok := inst.Variables["originIp"].([]byte)
			if ok {
				inst.Variables["originIp"] = net.IP(ipval).String()
			}

			if err := linfo.tmpl.Execute(buf, inst.Variables); err != nil {
				_ = h.env.Logger().Errorf("failed to execute template for log '%s': %v", inst.Name, err)
				// proceeding anyways
			}
			payload := buf.String()
			pool.PutBuffer(buf)

			if err := h.paperTrailLogger.Log(payload); err != nil {
				return h.logger.Errorf("error while recording the log message for papertrail: %v", err)
			}
		}

		// loggly
		if h.logglyClient != nil {
			if err := h.logglyClient.Send(structs.Map(inst)); err != nil {
				return h.logger.Errorf("error while recording the log message for loggly: %v", err)
			}
		}
	}
	return nil
}

func (h *logHandler) close() error {
	var err error
	if h.paperTrailLogger != nil {
		err = h.paperTrailLogger.Close()
	}
	return err
}
