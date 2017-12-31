package papertrail

import (
	"fmt"

	"istio.io/istio/mixer/pkg/adapter"
)

type LoggerImpl struct{}

func (l *LoggerImpl) VerbosityLevel(level adapter.VerbosityLevel) bool {
	return false
}
func (l *LoggerImpl) Infof(format string, args ...interface{}) {
	fmt.Printf("INFO: "+format+"\n", args...)
}
func (l *LoggerImpl) Warningf(format string, args ...interface{}) {
	fmt.Printf("WARN: "+format+"\n", args...)
}
func (l *LoggerImpl) Errorf(format string, args ...interface{}) error {
	err := fmt.Errorf("Error: "+format+"\n", args...)
	fmt.Printf("%v", err)
	return err
}
