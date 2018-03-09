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
	"strings"
	"time"

	"istio.io/istio/mixer/adapter/solarwinds/config"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/metric"

	"github.com/appoptics/appoptics-api-go"
)

type metricsHandlerInterface interface {
	handleMetric(context.Context, []*metric.Instance) error
	close() error
}

type metricsHandler struct {
	logger           adapter.Logger
	metricInfo       map[string]*config.Params_MetricInfo
	aoClient         *appoptics.Client
	aoBatchPersister *appoptics.BatchPersister
}

func newMetricsHandler(ctx context.Context, env adapter.Env, cfg *config.Params) (metricsHandlerInterface, error) {
	var lc *appoptics.Client
	var bp *appoptics.BatchPersister
	if strings.TrimSpace(cfg.AppopticsAccessToken) != "" {
		lc = appoptics.NewClient(cfg.AppopticsAccessToken)
		bp = appoptics.NewBatchPersister(lc.MeasurementsService(), true)
		env.ScheduleDaemon(bp.BatchAndPersistMeasurementsForever)
	}
	return &metricsHandler{
		logger:           env.Logger(),
		aoClient:         lc,
		aoBatchPersister: bp,
		metricInfo:       cfg.Metrics,
	}, nil
}

func (h *metricsHandler) handleMetric(_ context.Context, vals []*metric.Instance) error {
	measurements := []appoptics.Measurement{}
	for _, val := range vals {
		if mInfo, ok := h.metricInfo[val.Name]; ok {

			m := appoptics.Measurement{
				Name:  val.Name,
				Value: val.Value,
				Time:  time.Now().Unix(),
				Tags:  map[string]string{},
			}

			for _, label := range mInfo.LabelNames {
				// val.Dimensions[label] should exists because we have validated this before during config time.
				m.Tags[label] = adapter.Stringify(val.Dimensions[label])
			}
			measurements = append(measurements, m)
		}
	}
	if h.aoClient != nil {
		h.aoBatchPersister.MeasurementsSink() <- measurements
	}
	return nil
}

func (h *metricsHandler) close() error {
	if h.aoClient != nil {
		h.aoBatchPersister.MeasurementsStopBatchingChannel() <- true
	}
	return nil
}
