package appoptics

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"istio.io/istio/mixer/adapter/appoptics/appoptics"
	"istio.io/istio/mixer/adapter/appoptics/config"
	"istio.io/istio/mixer/adapter/appoptics/promadapter"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/metric"
)

type metricsHandlerInterface interface {
	HandleMetric(context.Context, []*metric.Instance) error
	Close() error
}

type metricsHandler struct {
	logger   adapter.Logger
	prepChan chan []*appoptics.Measurement

	stopChan chan struct{}
	errChan  chan error
	pushChan chan []*appoptics.Measurement
}

func NewMetricsHandler(ctx context.Context, env adapter.Env, cfg *config.Params) (metricsHandlerInterface, error) {
	env.Logger().Infof("AO - Invoking metrics handler build.")

	var err error
	// prepChan holds groups of Measurements to be batched
	prepChan := make(chan []*appoptics.Measurement)

	// pushChan holds groups of Measurements conforming to the size constraint described
	// by AppOptics.MeasurementPostMaxBatchSize
	pushChan := make(chan []*appoptics.Measurement)

	var stopChan = make(chan struct{})

	// errorChan is used to track persistence errors and shutdown when too many are seen
	errorChan := make(chan error)

	if strings.TrimSpace(cfg.AppopticsAccessToken) != "" {
		lc := appoptics.NewClient(cfg.AppopticsAccessToken, env.Logger())

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

	return &metricsHandler{
		logger:   env.Logger(),
		prepChan: prepChan,
		stopChan: stopChan,
		errChan:  errorChan,
		pushChan: pushChan,
	}, err
}

func (h *metricsHandler) HandleMetric(_ context.Context, vals []*metric.Instance) error {
	h.logger.Infof("AO - In the metrics handler. Received metrics: %#v", vals)
	measurements := []*appoptics.Measurement{}
	for _, val := range vals {
		h.logger.Infof("AO - In the metrics handler. Evaluating metric: %#v", val)
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

func (h *metricsHandler) Close() error {
	h.logger.Infof("AO - closing metrics handler")
	go func() { h.stopChan <- struct{}{} }()
	time.Sleep(time.Millisecond)
	close(h.prepChan)
	close(h.pushChan)
	close(h.errChan)
	close(h.stopChan)

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
