package appoptics

import (
	"context"
	"testing"

	"istio.io/istio/mixer/adapter/appoptics/config"
	"istio.io/istio/mixer/adapter/appoptics/papertrail"
)

func TestNewLogHandler(t *testing.T) {
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
			ctx := context.Background()

			lh, err := NewLogHandler(ctx, &adapterEnvInst{}, test.cfg)
			if err != nil {
				t.Errorf("Unexpected error: %v while running test: %s", err, t.Name())
				return
			}

			if !test.compareFn(lh) {
				t.Errorf("Unexpected response from compare function while running test: %s", t.Name())
			}
			logger.Infof("Finished %s - test run. . .", t.Name())
		})
	}
}
