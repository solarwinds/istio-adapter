package promadapter

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"istio.io/istio/mixer/adapter/appoptics/appoptics"
	"istio.io/istio/mixer/adapter/appoptics/papertrail"
)

func TestBatchMeasurements(t *testing.T) {

	t.Run("All Good", func(t *testing.T) {
		logger := &papertrail.LoggerImpl{}
		logger.Infof("Starting %s - test run. . .", t.Name())
		prepChan := make(chan []*appoptics.Measurement)
		pushChan := make(chan []*appoptics.Measurement)
		stopChan := make(chan struct{})

		go BatchMeasurements(prepChan, pushChan, stopChan, logger)

		go func() {
			measurements := []*appoptics.Measurement{}
			for i := 0; i < appoptics.MeasurementPostMaxBatchSize+1; i++ {
				measurements = append(measurements, &appoptics.Measurement{})
			}
			prepChan <- measurements
			close(prepChan)
			close(pushChan)
		}()
		count := 0
		for range pushChan {
			count++
		}
		if count != 1 {
			t.Errorf("Batching is not working properly. Expected batches is 1 but got %d", count)
		}
		close(stopChan)
		logger.Infof("Finished %s - test run. . .", t.Name())
	})

	t.Run("Using stop chan", func(t *testing.T) {
		logger := &papertrail.LoggerImpl{}
		logger.Infof("Starting %s - test run. . .", t.Name())
		prepChan := make(chan []*appoptics.Measurement)
		pushChan := make(chan []*appoptics.Measurement)
		stopChan := make(chan struct{})

		go func() {
			time.Sleep(time.Millisecond)
			stopChan <- struct{}{}
		}()
		BatchMeasurements(prepChan, pushChan, stopChan, logger)
		close(stopChan)
		close(prepChan)
		close(pushChan)
		logger.Infof("Finished %s - test run. . .", t.Name())
	})

}

type MockServiceAccessor struct {
	// MeasurementsService implements an interface for dealing with  Measurements
	MockMeasurementsService func() appoptics.MeasurementsCommunicator
}

func (s *MockServiceAccessor) MeasurementsService() appoptics.MeasurementsCommunicator {
	return s.MockMeasurementsService()
}

func TestPersistBatches(t *testing.T) {
	tests := []struct {
		name           string
		expectedCount  int64
		response       *http.Response
		error          error
		sendOnStopChan bool
	}{
		{
			name:          "Persist all good",
			expectedCount: 0,
			response: &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
			},
			error:          nil,
			sendOnStopChan: false,
		},
		{
			name:           "Response error",
			expectedCount:  1,
			response:       nil,
			error:          fmt.Errorf("Damn"),
			sendOnStopChan: false,
		},
		{
			name:           "Stop chan test",
			expectedCount:  0,
			response:       nil,
			error:          nil,
			sendOnStopChan: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := &papertrail.LoggerImpl{}
			logger.Infof("Starting %s - test run. . .\n", t.Name())
			pushChan := make(chan []*appoptics.Measurement)
			stopChan := make(chan struct{})
			errChan := make(chan error)
			var count int64
			if test.sendOnStopChan {
				go func() {
					time.Sleep(time.Millisecond)
					stopChan <- struct{}{}
				}()
			}
			go func() {
				time.Sleep(50 * time.Millisecond)
				pushChan <- []*appoptics.Measurement{
					{}, {}, {},
				}
			}()
			go func() {
				time.Sleep(time.Millisecond)
				<-errChan
				atomic.AddInt64(&count, 1)
			}()
			go PersistBatches(&MockServiceAccessor{
				MockMeasurementsService: func() appoptics.MeasurementsCommunicator {
					return &appoptics.MockMeasurementsService{
						OnCreate: func(measurements []*appoptics.Measurement) (*http.Response, error) {
							return test.response, test.error
						},
					}
				},
			}, pushChan, stopChan, errChan, logger)
			time.Sleep(2 * time.Second)
			if atomic.LoadInt64(&count) != test.expectedCount {
				t.Errorf("Count did not match the expected count: %d", test.expectedCount)
			}
			logger.Infof("Closing channels. . .")
			close(pushChan)
			close(stopChan)
			close(errChan)
			logger.Infof("Finished %s - test run. . .", t.Name())
		})
	}
}
