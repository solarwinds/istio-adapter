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

// Package prometheus publishes metric values collected by Mixer for
// ingestion by prometheus.

package appoptics

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"istio.io/istio/mixer/adapter/appoptics/config"
	"istio.io/istio/mixer/adapter/appoptics/papertrail"

	"istio.io/istio/mixer/adapter/appoptics/appoptics"
	"istio.io/istio/mixer/adapter/appoptics/promadapter"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/logentry"
	"istio.io/istio/mixer/template/metric"
)

type (
	builder struct {
		cfg *config.Params
	}

	handler struct {
		logger           adapter.Logger
		prepChan         chan []*appoptics.Measurement
		paperTrailLogger *papertrail.PaperTrailLogger
	}
)

var (
	charReplacer = strings.NewReplacer("/", "_", ".", "_", " ", "_", "-", "")

	_ metric.HandlerBuilder = &builder{}
	_ metric.Handler        = &handler{}
)

const (
	namespace = "istio"
)

// GetInfo returns the Info associated with this adapter.
func GetInfo() adapter.Info {
	singletonBuilder := &builder{}
	singletonBuilder.clearState()
	return adapter.Info{
		Name:        "appoptics",
		Impl:        "istio.io/istio/mixer/adapter/appoptics",
		Description: "Publishes metrics to appoptics and logs to papertrail",
		SupportedTemplates: []string{
			metric.TemplateName,
			logentry.TemplateName,
		},
		NewBuilder:    func() adapter.HandlerBuilder { return singletonBuilder },
		DefaultConfig: &config.Params{},
	}
}

func (b *builder) clearState() {
}

func (b *builder) SetMetricTypes(map[string]*metric.Type) {}
func (b *builder) SetAdapterConfig(cfg adapter.Config) {
	// this is a common adapter config for both log and metric
	b.cfg = cfg.(*config.Params)
}

func (b *builder) SetLogEntryTypes(entries map[string]*logentry.Type) {

}
func (b *builder) Validate() *adapter.ConfigErrors { return nil }
func (b *builder) Build(ctx context.Context, env adapter.Env) (adapter.Handler, error) {
	env.Logger().Infof("AO - Invoking AO build.")

	var err error
	// prepChan holds groups of Measurements to be batched
	prepChan := make(chan []*appoptics.Measurement)

	if strings.TrimSpace(b.cfg.AppopticsAccessToken) != "" {
		lc := appoptics.NewClient(b.cfg.AppopticsAccessToken, env.Logger())

		// pushChan holds groups of Measurements conforming to the size constraint described
		// by AppOptics.MeasurementPostMaxBatchSize
		pushChan := make(chan []*appoptics.Measurement)

		var stopChan = make(chan bool)

		// errorChan is used to track persistence errors and shutdown when too many are seen
		errorChan := make(chan error)

		go promadapter.BatchMeasurements(prepChan, pushChan, stopChan, env.Logger())
		go promadapter.PersistBatches(lc, pushChan, stopChan, errorChan, env.Logger())
		go promadapter.ManagePersistenceErrors(errorChan, stopChan, env.Logger())
	} else {
		go func() {
			// to drain the channel
			for range prepChan {

			}
		}()
	}

	var pp *papertrail.PaperTrailLogger
	if strings.TrimSpace(b.cfg.PapertrailUrl) != "" {
		pp, err = papertrail.NewPaperTrailLogger(b.cfg.PapertrailUrl, b.cfg.PapertrailLocalRetention, b.cfg.Logs, env.Logger())
	}
	return &handler{env.Logger(), prepChan, pp}, err
}

func (h *handler) HandleMetric(_ context.Context, vals []*metric.Instance) error {
	h.logger.Infof("AO - In the handler. Received metrics: %#v", vals)
	measurements := []*appoptics.Measurement{}
	for _, val := range vals {
		h.logger.Infof("AO - In the handler. Evaluating metric: %#v", val)
		h.logger.Infof("Received Metric Name: %s, Dimensions: %v, Value: %v", val.Name, val.Dimensions, val.Value)
		var merticVal float64
		merticVal = h.aoVal(val.Value)

		m := &appoptics.Measurement{
			Name:  val.Name,
			Value: merticVal,
			Time:  time.Now().Unix(),
			Tags:  appoptics.MeasurementTags{},
		}
		for k, v := range val.Dimensions {
			switch v.(type) {
			case int, int32, int64:
				m.Tags[k] = fmt.Sprintf("%d", v)
			case float64:
				m.Tags[k] = fmt.Sprintf("%f", v)
			default:
				m.Tags[k], _ = v.(string)
			}
		}

		measurements = append(measurements, m)
	}
	h.prepChan <- measurements

	return nil
}

func (h *handler) HandleLogEntry(ctx context.Context, values []*logentry.Instance) error {
	h.logger.Infof("AO - In the log handler")
	for _, inst := range values {
		if h.paperTrailLogger != nil {
			err := h.paperTrailLogger.Log(inst)
			if err != nil {
				h.logger.Errorf("AO - log error: %v", err)
				return err
			}
		}
	}
	return nil
}

func (h *handler) Close() error {
	var err error
	// h.logger.Infof("AO - closing handler")
	h.logger.Infof("AO - closing handler")
	if h.paperTrailLogger != nil {
		err = h.paperTrailLogger.Close()
		// return h.srv.Close()
	}
	return err
}

func (h *handler) aoVal(i interface{}) float64 {
	switch vv := i.(type) {
	case float64:
		return vv
	case int64:
		return float64(vv)
	case time.Duration:
		// use seconds for now
		return vv.Seconds()
	case string:
		f, err := strconv.ParseFloat(vv, 64)
		if err != nil {
			h.logger.Infof("AO - Error parsing metric val: %v", vv)
			// return math.NaN(), err
			f = 0
		}
		return f
	default:
		// return math.NaN(), fmt.Errorf("could not extract numeric value for %v", val)
		h.logger.Infof("AO - could not extract numeric value for %v", vv)
		return 0
	}
}
