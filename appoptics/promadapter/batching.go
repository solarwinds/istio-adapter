package promadapter

import (
	"net/http"
	"time"

	"bytes"

	"istio.io/istio/mixer/adapter/appoptics/appoptics"
	"istio.io/istio/mixer/pkg/adapter"
)

// BatchMeasurements reads slices of AppOptics.Measurement types off a channel populated by the web handler
// and packages them into batches conforming to the limitations imposed by the API.
func BatchMeasurements(prepChan <-chan []*appoptics.Measurement, pushChan chan<- []*appoptics.Measurement, stopChan <-chan bool, logger adapter.Logger) {
	var currentBatch []*appoptics.Measurement
	for {
		select {
		case mslice := <-prepChan:
			logger.Infof("AO - batching measurements: %v", mslice)
			currentBatch = append(currentBatch, mslice...)
			logger.Infof("AO - current batch size: %d", len(currentBatch))
			logger.Infof("AO - batch max size: %d", appoptics.MeasurementPostMaxBatchSize)
			if len(currentBatch) >= appoptics.MeasurementPostMaxBatchSize {
				pushBatch := currentBatch[:appoptics.MeasurementPostMaxBatchSize]
				pushChan <- pushBatch
				currentBatch = currentBatch[appoptics.MeasurementPostMaxBatchSize:]
			}
		case <-stopChan:
			break
		}
	}
}

// PersistBatches reads maximal slices of AppOptics.Measurement types off a channel and persists them to the remote AppOptics
// API. Errors are placed on the error channel.
func PersistBatches(lc appoptics.ServiceAccessor, pushChan <-chan []*appoptics.Measurement, stopChan <-chan bool, errorChan chan<- error, logger adapter.Logger) {
	ticker := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-ticker.C:
			batch := <-pushChan
			logger.Infof("AO - persisting batch. . .")
			err := persistBatch(lc, batch, logger)
			if err != nil {
				errorChan <- err
			}
		case <-stopChan:
			ticker.Stop()
			break
		}
	}
}

// ManagePersistenceErrors tracks errors on the provided channel and sends a stop signal if the ErrorLimit is reached
func ManagePersistenceErrors(errorChan <-chan error, stopChan chan<- bool, logger adapter.Logger) {
	// var errors []error
	for {
		select {
		case err := <-errorChan:
			// errors = append(errors, err)
			// if len(errors) > config.PushErrorLimit() {
			// 	stopChan <- true
			// 	break
			// }
			logger.Errorf("AO - Persistence Errors: %v", err)
		}

	}
}

// persistBatch sends to the remote AppOptics endpoint unless config.SendStats() returns false, when it prints to stdout
func persistBatch(lc appoptics.ServiceAccessor, batch []*appoptics.Measurement, logger adapter.Logger) error {
	logger.Infof("AO - persisting %d Measurements to AppOptics", len(batch))
	resp, err := lc.MeasurementsService().Create(batch)
	if resp == nil {
		logger.Infof("AO - response is nil")
		return err
	}
	dumpResponse(resp, logger)
	return nil
}
func dumpResponse(resp *http.Response, logger adapter.Logger) {
	buf := new(bytes.Buffer)
	logger.Infof("AO - response status: %s", resp.Status)
	if resp.Body != nil {
		buf.ReadFrom(resp.Body)
		logger.Infof("AO - response body: %s", string(buf.Bytes()))
	}
}