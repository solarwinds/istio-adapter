package papertrail

import (
	"reflect"
	"testing"
	"time"

	"istio.io/istio/mixer/adapter/appoptics/config"
	"istio.io/istio/mixer/pkg/adapter"
)

type loggerImpl struct{}

func (l *loggerImpl) VerbosityLevel(level adapter.VerbosityLevel) bool {
	return false
}
func (l *loggerImpl) Infof(format string, args ...interface{})        {}
func (l *loggerImpl) Warningf(format string, args ...interface{})     {}
func (l *loggerImpl) Errorf(format string, args ...interface{}) error { return nil }

func TestNewPaperTrailLogger(t *testing.T) {
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
				logger:        &loggerImpl{},
			},
			want: &PaperTrailLogger{
				paperTrailURL:   "hello.world.org",
				retentionPeriod: 24 * time.Hour,
				log:             &loggerImpl{},
				logInfos:        map[string]*logInfo{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPaperTrailLogger(tt.args.paperTrailURL, tt.args.logRetentionStr, tt.args.logConfigs, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPaperTrailLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPaperTrailLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}
