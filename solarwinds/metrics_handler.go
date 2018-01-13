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

package solarwinds

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"istio.io/istio/mixer/adapter/solarwinds/config"
	"istio.io/istio/mixer/adapter/solarwinds/internal/appoptics"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/metric"
)

// MeasurementPostMaxBatchSize defines the max number of Measurements to send to the API at once
const MeasurementPostMaxBatchSize = 1000

type metricsHandlerInterface interface {
	handleMetric(context.Context, []*metric.Instance) error
	close() error
}

type metricsHandler struct {
	logger   adapter.Logger
	prepChan chan []*appoptics.Measurement

	stopChan chan struct{}
	pushChan chan []*appoptics.Measurement

	loopFactor *bool
}

func newMetricsHandler(ctx context.Context, env adapter.Env, cfg *config.Params) (metricsHandlerInterface, error) {
	buffChanSize := runtime.NumCPU() * 10

	loopFactor := true

	var err error
	// prepChan holds groups of Measurements to be batched
	prepChan := make(chan []*appoptics.Measurement, buffChanSize)

	// pushChan holds groups of Measurements conforming to the size constraint described
	// by AppOptics.MeasurementPostMaxBatchSize
	pushChan := make(chan []*appoptics.Measurement, buffChanSize)

	var stopChan = make(chan struct{})

	if strings.TrimSpace(cfg.AppopticsAccessToken) != "" {
		lc := appoptics.NewClient(cfg.AppopticsAccessToken, env.Logger())

		batchSize := cfg.AppopticsBatchSize
		if batchSize <= 0 || batchSize > MeasurementPostMaxBatchSize {
			batchSize = MeasurementPostMaxBatchSize
		}

		env.ScheduleDaemon(func() {
			appoptics.BatchMeasurements(&loopFactor, prepChan, pushChan, stopChan, int(batchSize), env.Logger())
		})
		env.ScheduleDaemon(func() {
			appoptics.PersistBatches(&loopFactor, lc, pushChan, stopChan, env.Logger())
		})
	} else {
		env.ScheduleDaemon(func() {
			// to drain the channel
			for range prepChan {

			}
		})
	}

	return &metricsHandler{
		logger:     env.Logger(),
		prepChan:   prepChan,
		stopChan:   stopChan,
		pushChan:   pushChan,
		loopFactor: &loopFactor,
	}, err
}

func (h *metricsHandler) handleMetric(_ context.Context, vals []*metric.Instance) error {
	measurements := []*appoptics.Measurement{}
	for _, val := range vals {
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

func (h *metricsHandler) close() error {
	close(h.prepChan)
	close(h.pushChan)
	close(h.stopChan)
	*h.loopFactor = false

	return nil
}

func (h *metricsHandler) aoVal(i interface{}) float64 {
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
			h.logger.Errorf("ao - Error parsing metric val: %v", vv)
			// return math.NaN(), err
			f = 0
		}
		return f
	default:
		// return math.NaN(), fmt.Errorf("could not extract numeric value for %v", val)
		h.logger.Errorf("ao - could not extract numeric value for %v", vv)
		return 0
	}
}
