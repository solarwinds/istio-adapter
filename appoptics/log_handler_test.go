package appoptics

import (
	"context"
	"fmt"
	"testing"

	"istio.io/istio/mixer/adapter/appoptics/config"
	"istio.io/istio/mixer/adapter/appoptics/papertrail"
	"istio.io/istio/mixer/template/logentry"
)

func TestNewLogHandler(t *testing.T) {
	ctx := context.Background()

	type testData struct {
		name      string
		cfg       *config.Params
		compareFn func(logHandlerInterface) bool
	}
	tests := []*testData{
		&testData{
			name: "All good",
			cfg: &config.Params{
				PapertrailUrl: "hello.world.org",
			},
			compareFn: func(lhi logHandlerInterface) bool {
				lh, _ := lhi.(*logHandler)
				return lh.paperTrailLogger != nil
			},
		},
		&testData{
			name: "Empty ref",
			cfg:  &config.Params{},
			compareFn: func(lhi logHandlerInterface) bool {
				lh, _ := lhi.(*logHandler)
				pp, _ := lh.paperTrailLogger.(*papertrail.PaperTrailLogger)
				return pp == nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := &papertrail.LoggerImpl{}
			logger.Infof("Starting %s - test run. . .", t.Name())
			defer logger.Infof("Finished %s - test run. . .", t.Name())

			lh, err := NewLogHandler(ctx, &adapterEnvInst{}, test.cfg)
			if err != nil {
				t.Errorf("Unexpected error: %v while running test: %s", err, t.Name())
				return
			}

			if !test.compareFn(lh) {
				t.Errorf("Unexpected response from compare function while running test: %s", t.Name())
			}
		})
	}
}

func TestHandleLogEntry(t *testing.T) {
	ctx := context.Background()
	logger := &papertrail.LoggerImpl{}
	t.Run("All good", func(t *testing.T) {
		logger.Infof("Starting %s - test run. . .", t.Name())
		defer logger.Infof("Finished %s - test run. . .", t.Name())
		port := 34543
		serverStopChan := make(chan struct{})
		serverTrackChan := make(chan struct{})
		go papertrail.RunUDPServer(port, logger, serverStopChan, serverTrackChan)
		go func() {
			count := 0
			for range serverTrackChan {
				count++
			}
			if count != 1 {
				t.Errorf("Expected data count (1) received by server dont match the actual number: %d", count)
			}
		}()

		lh, err := NewLogHandler(ctx, &adapterEnvInst{}, &config.Params{
			PapertrailUrl: fmt.Sprintf("localhost:%d", port),
			Logs: []*config.Params_LogInfo{
				{
					InstanceName: "params1",
				},
			},
		})
		err = lh.HandleLogEntry(ctx, []*logentry.Instance{
			&logentry.Instance{
				Name:      "params1",
				Variables: map[string]interface{}{},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error while executing test: %s - err: %v", t.Name(), err)
			return
		}
	})

	t.Run("papertrail instance is nil", func(t *testing.T) {
		logger.Infof("Starting %s - test run. . .", t.Name())
		lh, err := NewLogHandler(ctx, &adapterEnvInst{}, &config.Params{})
		if err != nil {
			t.Errorf("Unexpected error while executing test: %s - err: %v", t.Name(), err)
			return
		}
		err = lh.HandleLogEntry(ctx, []*logentry.Instance{
			&logentry.Instance{},
		})
		if err != nil {
			t.Errorf("Unexpected error while executing test: %s - err: %v", t.Name(), err)
			return
		}
	})
}

/*
func TestLog(t *testing.T) {
	logger := &LoggerImpl{}
	t.Run("No log info for msg name", func(t *testing.T) {
		logger.Infof("Starting TestLog - %s - test run. . .", t.Name())
		pp := &PaperTrailLogger{
			paperTrailURL: "hello.world.hey",
			log:           &LoggerImpl{},
			logInfos:      map[string]*logInfo{},
		}

		if pp.Log(&logentry.Instance{
			Name: "NO ENTRY",
		}) == nil {
			t.Error("An error is expected here.")
		}
		logger.Infof("Finished TestLog - %s - test run. . .", t.Name())
	})

	t.Run("All Good", func(t *testing.T) {
		logger.Infof("Starting %s - test run. . .", t.Name())
		port := 6767
		urlLocal := "localhost"

		ppi, err := NewPaperTrailLogger(fmt.Sprintf("%s:%d", urlLocal, port), "1h", []*config.Params_LogInfo{
			{
				InstanceName: "params1",
			},
		}, logger)
		if err != nil {
			t.Errorf("No error was expected")
		}

		pp, _ := ppi.(*PaperTrailLogger)

		pcount := getKeyCount(pp)

		serverStopChan := make(chan struct{})
		serverTrackChan := make(chan struct{})
		go runUDPServer(port, logger, serverStopChan, serverTrackChan)
		go func() {
			count := 0
			for range serverTrackChan {
				count++
			}
			if count != 1 {
				t.Errorf("Expected data count (1) received by server dont match the actual number: %d", count)
			}
		}()

		if err = pp.Log(&logentry.Instance{
			Name:      "params1",
			Variables: map[string]interface{}{},
		}); err != nil {
			t.Errorf("No error was expected")
		}

		count := getKeyCount(pp)
		if count-pcount != 1 {
			t.Error("key counts dont match")
		}
		time.Sleep(2 * time.Second)
		count = getKeyCount(pp)
		if count-pcount != 0 {
			t.Error("key counts dont match")
		}

		serverStopChan <- struct{}{}
		close(serverStopChan)
		close(serverTrackChan)

		logger.Infof("Finished %s - test run. . .", t.Name())

	})
}
*/
