package papertrail

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/boltdb/bolt"
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
			logger.Infof("Starting TestNewPaperTrailLogger - %s - test run. . .", tt.name)

			got, err := NewPaperTrailLogger(tt.args.paperTrailURL, tt.args.logRetentionStr, tt.args.logConfigs, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPaperTrailLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("Expected a non-nil instance")
			}
			logger.Infof("Finished TestNewPaperTrailLogger - %s - test run. . .", tt.name)
		})
	}
}

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

func getKeyCount(pp *PaperTrailLogger) int {
	count := 0
	pp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b != nil {
			c := b.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				count++
			}
		}
		return nil
	})
	return count
}

func runUDPServer(port int, logger adapter.Logger, stopChan chan struct{}, trackChan chan struct{}) {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Errorf("Error: %v", err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logger.Errorf("Error: %v", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	go func() {
		for {
			noOfBytes, remoteAddr, err := conn.ReadFromUDP(buf)
			logger.Infof("udp server - data received: %s from %v", strings.TrimSpace(string(buf[0:noOfBytes])), remoteAddr)
			if err != nil {
				logger.Errorf("Error: %v", err)
				return
			}
			trackChan <- struct{}{}
		}
	}()
	var tobrk bool
	for {
		select {
		case <-stopChan:
			tobrk = true
		default:
		}
		if tobrk {
			break
		}
	}
}
