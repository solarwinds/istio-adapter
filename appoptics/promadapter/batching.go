package promadapter

import (
	"net/http"
	"time"

	"bytes"

	"istio.io/istio/mixer/adapter/appoptics/appoptics"
	"istio.io/istio/mixer/adapter/appoptics/config"
	"istio.io/istio/mixer/pkg/adapter"
)

// BatchMeasurements reads slices of AppOptics.Measurement types off a channel populated by the web handler
// and packages them into batches conforming to the limitations imposed by the API.
func BatchMeasurements(loopFactor *bool, prepChan <-chan []*appoptics.Measurement, pushChan chan<- []*appoptics.Measurement, stopChan <-chan struct{}, logger adapter.Logger) {
	var currentBatch []*appoptics.Measurement
	dobrk := false
	ticker := time.NewTicker(time.Millisecond * 500)
	for *loopFactor {
		select {
		case mslice := <-prepChan:
			if logger.VerbosityLevel(config.DebugLevel) {
				logger.Infof("AO - batching measurements: %v", mslice)
			}
			currentBatch = append(currentBatch, mslice...)
			if logger.VerbosityLevel(config.DebugLevel) {
				logger.Infof("AO - current batch size: %d, batch max size: %d", len(currentBatch), appoptics.MeasurementPostMaxBatchSize)
			}
			if len(currentBatch) >= appoptics.MeasurementPostMaxBatchSize {
				pushBatch := currentBatch[:appoptics.MeasurementPostMaxBatchSize]
				pushChan <- pushBatch
				currentBatch = currentBatch[appoptics.MeasurementPostMaxBatchSize:]
			}
		case <-ticker.C: // to drain based on time as well
			if len(currentBatch) > 0 {
				if len(currentBatch) >= appoptics.MeasurementPostMaxBatchSize {
					pushBatch := currentBatch[:appoptics.MeasurementPostMaxBatchSize]
					pushChan <- pushBatch
					currentBatch = currentBatch[appoptics.MeasurementPostMaxBatchSize:]
				} else {
					pushChan <- currentBatch
					currentBatch = []*appoptics.Measurement{}
				}
			}
		case <-stopChan:
			dobrk = true
		}
		if dobrk {
			break
		}
	}
}

// PersistBatches reads maximal slices of AppOptics.Measurement types off a channel and persists them to the remote AppOptics
// API. Errors are placed on the error channel.
func PersistBatches(loopFactor *bool, lc appoptics.ServiceAccessor, pushChan <-chan []*appoptics.Measurement, stopChan <-chan struct{}, errorChan chan<- error, logger adapter.Logger) {
	ticker := time.NewTicker(time.Millisecond * 500)
	dobrk := false
	for *loopFactor {
		select {
		case <-ticker.C:
			batch := <-pushChan
			if logger.VerbosityLevel(config.DebugLevel) {
				logger.Infof("AO - persisting batch. . .")
			}
			err := persistBatch(lc, batch, logger)
			if err != nil {
				errorChan <- err
			}
		case <-stopChan:
			ticker.Stop()
			dobrk = true
		}
		if dobrk {
			break
		}
	}
}

// ManagePersistenceErrors tracks errors on the provided channel and sends a stop signal if the ErrorLimit is reached
func ManagePersistenceErrors(loopFactor *bool, errorChan <-chan error, stopChan chan<- struct{}, logger adapter.Logger) {
	// var errors []error
	for *loopFactor {
		select {
		case err := <-errorChan:
			if err != nil {
				// errors = append(errors, err)
				// if len(errors) > config.PushErrorLimit() {
				// 	stopChan <- true
				// 	break
				// }
				logger.Errorf("AO - Persistence Errors: %v", err)
			}
		}

	}
}

// persistBatch sends to the remote AppOptics endpoint unless config.SendStats() returns false, when it prints to stdout
func persistBatch(lc appoptics.ServiceAccessor, batch []*appoptics.Measurement, logger adapter.Logger) error {
	if logger.VerbosityLevel(config.DebugLevel) {
		logger.Infof("AO - persisting %d Measurements to AppOptics", len(batch))
	}
	if len(batch) > 0 {
		resp, err := lc.MeasurementsService().Create(batch)
		if err != nil {
			logger.Errorf("AO - persist error: %v", err)
			return err
		}
		dumpResponse(resp, logger)
	}
	return nil
}

func dumpResponse(resp *http.Response, logger adapter.Logger) {
	buf := new(bytes.Buffer)
	if logger.VerbosityLevel(config.DebugLevel) {
		logger.Infof("AO - response status: %s", resp.Status)
	}
	if resp.Body != nil {
		buf.ReadFrom(resp.Body)
		if logger.VerbosityLevel(config.DebugLevel) {
			logger.Infof("AO - response body: %s", string(buf.Bytes()))
		}
	}
}
