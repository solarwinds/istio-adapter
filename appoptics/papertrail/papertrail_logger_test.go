package papertrail

import (
	"fmt"
	"testing"
	"time"

	"istio.io/istio/mixer/adapter/appoptics/config"
	"istio.io/istio/mixer/pkg/adapter"

	"istio.io/istio/mixer/template/logentry"
)

func TestNewPaperTrailLogger(t *testing.T) {
	logger := &LoggerImpl{}
	type (
		args struct {
			paperTrailURL   string
			logRetentionStr string
			logConfigs      []*config.Params_LogInfo
			logger          adapter.Logger
		}
		testData struct {
			name    string
			args    args
			want    *PaperTrailLogger
			wantErr bool
		}
	)
	tests := []testData{
		{
			name: "All good",
			args: args{
				paperTrailURL: "hello.world.org",
				logger:        &LoggerImpl{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.Infof("Starting %s - test run. . .", t.Name())
			defer logger.Infof("Finished %s - test run. . .", t.Name())

			got, err := NewPaperTrailLogger(tt.args.paperTrailURL, tt.args.logRetentionStr, tt.args.logConfigs, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPaperTrailLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("Expected a non-nil instance")
			}

		})
	}
}

func TestLog(t *testing.T) {
	logger := &LoggerImpl{}
	loopFactor := true
	t.Run("No log info for msg name", func(t *testing.T) {
		logger.Infof("Starting %s - test run. . .", t.Name())
		defer logger.Infof("Finished %s - test run. . .", t.Name())

		pp := &PaperTrailLogger{
			paperTrailURL: "hello.world.hey",
			log:           &LoggerImpl{},
			logInfos:      map[string]*logInfo{},
			loopFactor:    loopFactor,
		}

		if pp.Log(&logentry.Instance{
			Name: "NO ENTRY",
		}) == nil {
			t.Error("An error is expected here.")
		}
	})

	t.Run("All Good", func(t *testing.T) {
		logger.Infof("Starting %s - test run. . .", t.Name())
		defer logger.Infof("Finished %s - test run. . .", t.Name())
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
		go RunUDPServer(port, logger, serverStopChan, serverTrackChan)
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
	})
}

func getKeyCount(pp *PaperTrailLogger) int {
	count := 0
	pp.cmap.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	return count
}
