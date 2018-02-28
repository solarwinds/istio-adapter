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

package appoptics

import (
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	test2 "istio.io/istio/mixer/pkg/adapter/test"
)

func TestBatchMeasurements(t *testing.T) {

	if os.Getenv("RACE_TEST") == "true" {
		t.Skip("Test is broken for race testing, see issue #3207")
	}
	t.Run("All Good", func(t *testing.T) {
		env := test2.NewEnv(t)
		logger := env.Logger()

		logger.Infof("Starting %s - test run. . .", t.Name())
		defer logger.Infof("Finished %s - test run. . .", t.Name())
		prepChan := make(chan []*Measurement)
		pushChan := make(chan []*Measurement)
		stopChan := make(chan struct{})

		loopFactor := true
		batchSize := 100

		go BatchMeasurements(&loopFactor, prepChan, pushChan, stopChan, batchSize, logger)

		go func() {
			measurements := []*Measurement{}
			for i := 0; i < batchSize+1; i++ {
				measurements = append(measurements, &Measurement{})
			}
			prepChan <- measurements
			loopFactor = false
			time.Sleep(time.Millisecond)
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
	})

	t.Run("Using stop chan", func(t *testing.T) {
		env := test2.NewEnv(t)
		logger := env.Logger()
		logger.Infof("Starting %s - test run. . .", t.Name())
		defer logger.Infof("Finished %s - test run. . .", t.Name())
		prepChan := make(chan []*Measurement)
		pushChan := make(chan []*Measurement)
		stopChan := make(chan struct{})
		batchSize := 100
		loopFactor := true
		go func() {
			time.Sleep(time.Millisecond)
			stopChan <- struct{}{}
		}()
		BatchMeasurements(&loopFactor, prepChan, pushChan, stopChan, batchSize, logger)
		loopFactor = false
		close(prepChan)
		close(pushChan)
		close(stopChan)
	})

}

type MockServiceAccessor struct {
	// MeasurementsService implements an interface for dealing with  Measurements
	MockMeasurementsService func() MeasurementsCommunicator
}

func (s *MockServiceAccessor) MeasurementsService() MeasurementsCommunicator {
	return s.MockMeasurementsService()
}

func TestPersistBatches(t *testing.T) {
	if os.Getenv("RACE_TEST") == "true" {
		t.Skip("Test is broken for race testing, see issue #3209")
	}

	tests := []struct {
		name           string
		expectedCount  int32
		response       *http.Response
		error          error
		sendOnStopChan bool
	}{
		{
			name:          "Persist all good",
			expectedCount: 1,
			response: &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
			},
			error:          nil,
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
			env := test2.NewEnv(t)
			logger := env.Logger()
			logger.Infof("Starting %s - test run. . .\n", t.Name())
			defer logger.Infof("Finished %s - test run. . .", t.Name())
			pushChan := make(chan []*Measurement)
			stopChan := make(chan struct{})
			var count int32
			var wg sync.WaitGroup
			if test.sendOnStopChan {
				wg.Add(1)
				go func() {
					time.Sleep(time.Millisecond)
					stopChan <- struct{}{}
					wg.Done()
				}()
			} else {
				go func() {
					time.Sleep(50 * time.Millisecond)
					pushChan <- []*Measurement{
						{}, {}, {},
					}
				}()
			}
			loopFactor := true

			go PersistBatches(&loopFactor, &MockServiceAccessor{
				MockMeasurementsService: func() MeasurementsCommunicator {
					return &MockMeasurementsService{
						OnCreate: func(measurements []*Measurement) (*http.Response, error) {
							atomic.AddInt32(&count, 1)
							return test.response, test.error
						},
					}
				},
			}, pushChan, stopChan, logger)

			time.Sleep(2 * time.Second)
			logger.Infof("%s - waiting...\n", t.Name())
			if test.sendOnStopChan {
				wg.Wait()
			}
			if atomic.LoadInt32(&count) != test.expectedCount {
				t.Errorf("Count did not match the expected count: %d", test.expectedCount)
			}
			logger.Infof("Closing channels. . .")
			loopFactor = false
			close(pushChan)
			close(stopChan)
		})
	}
}
