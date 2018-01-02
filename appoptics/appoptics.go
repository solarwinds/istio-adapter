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

	"istio.io/istio/mixer/adapter/appoptics/config"

	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/logentry"
	"istio.io/istio/mixer/template/metric"
)

type (
	builder struct {
		cfg *config.Params
	}

	handler struct {
		logger         adapter.Logger
		metricsHandler metricsHandlerInterface
		logHandler     logHandlerInterface
		loopFactor     *bool
	}
)

//var (
//	charReplacer = strings.NewReplacer("/", "_", ".", "_", " ", "_", "-", "")

//	_ metric.HandlerBuilder = &builder{}
//	_ metric.Handler        = &handler{}
//)

//const (
//	namespace = "istio"
//)

// GetInfo returns the Info associated with this adapter.
func GetInfo() adapter.Info {
	return adapter.Info{
		Name:        "appoptics",
		Impl:        "istio.io/istio/mixer/adapter/appoptics",
		Description: "Publishes metrics to appoptics and logs to papertrail",
		SupportedTemplates: []string{
			metric.TemplateName,
			logentry.TemplateName,
		},
		NewBuilder:    func() adapter.HandlerBuilder { return &builder{} },
		DefaultConfig: &config.Params{},
	}
}

//func (b *builder) SetMetricTypes(map[string]*metric.Type) {}
func (b *builder) SetAdapterConfig(cfg adapter.Config) {
	// this is a common adapter config for both log and metric
	b.cfg = cfg.(*config.Params)
}

func (b *builder) SetMetricTypes(map[string]*metric.Type) {}

func (b *builder) SetLogEntryTypes(entries map[string]*logentry.Type) {}

func (b *builder) Validate() *adapter.ConfigErrors { return nil }

func (b *builder) Build(ctx context.Context, env adapter.Env) (adapter.Handler, error) {
	logger := env.Logger()

	loopFactor := true

	if logger.VerbosityLevel(config.DebugLevel) {
		logger.Infof("AO - Invoking AO build.")
	}

	m, err := NewMetricsHandler(ctx, env, b.cfg, &loopFactor)
	if err != nil {
		return nil, err
	}
	l, err := NewLogHandler(ctx, env, b.cfg, &loopFactor)
	if err != nil {
		return nil, err
	}
	return &handler{
		metricsHandler: m,
		logHandler:     l,
		logger:         env.Logger(),
		loopFactor:     &loopFactor,
	}, nil
}

func (h *handler) HandleMetric(ctx context.Context, vals []*metric.Instance) error {
	if h.logger.VerbosityLevel(config.DebugLevel) {
		h.logger.Infof("AO - In the metrics handler")
	}
	return h.metricsHandler.HandleMetric(ctx, vals)
}

func (h *handler) HandleLogEntry(ctx context.Context, values []*logentry.Instance) error {
	if h.logger.VerbosityLevel(config.DebugLevel) {
		h.logger.Infof("AO - In the log handler")
	}
	return h.logHandler.HandleLogEntry(ctx, values)
}

func (h *handler) Close() error {
	var err error
	if h.logger.VerbosityLevel(config.DebugLevel) {
		h.logger.Infof("AO - closing handler")
	}

	*h.loopFactor = false // to kill the loops

	if h.metricsHandler != nil {
		err = h.metricsHandler.Close()
		if err != nil {
			return err
		}
	}
	if h.logHandler != nil {
		err = h.logHandler.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
