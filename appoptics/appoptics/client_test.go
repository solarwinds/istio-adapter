package appoptics

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"istio.io/istio/mixer/adapter/appoptics/papertrail"
)

var paramErrorFixture = `{
    "measurements": {
        "summary": {
            "total": 6,
            "accepted": 0,
            "failed": 6
        }
    },
    "errors": [
        {
            "param": "time",
            "value": 1507056682391,
            "reason": "Is too far in the future (>30 minutes ahead). Check for local clock drift or enable NTP."
        }
    ]
}`

var requestErrorFixture = `{
  "errors": {
    "request": [
      "Please use secured connection through https!",
      "Please provide credentials for authentication."
    ]
  }
}`

// TODO: this probably represents a bug, as I got this from the same request that produced "paramErrorFixture" above
// but simply altered the URL scheme to `http` instead of `https`
var genericErrorSliceFixture = `{
    "errors": [
        "must specify metric name or compose parameter"
    ]
}`

var textErrorFixture = `Credentials are required to access this resource.`

func TestNewRequest(t *testing.T) {

	t.Run("All good", func(t *testing.T) {
		logger := &papertrail.LoggerImpl{}
		logger.Infof("Starting %s - test run. . .\n", t.Name())

		c := NewClient("some string", logger)
		req, err := c.NewRequest("POST", "measurements", &Measurement{})
		if req == nil || err != nil {
			t.Errorf("There was an error creating request")
		}

		logger.Infof("Finishing %s - test run. . .\n", t.Name())
	})

	t.Run("Request error", func(t *testing.T) {
		logger := &papertrail.LoggerImpl{}
		logger.Infof("Starting %s - test run. . .\n", t.Name())

		c := NewClient("some string", logger)
		req, err := c.NewRequest("POST", "measurements", make(chan int))
		if req != nil || err == nil {
			t.Errorf("There should have been an error creating request")
		}

		logger.Infof("Finishing %s - test run. . .\n", t.Name())
	})
}

func TestDo(t *testing.T) {
	t.Run("All good", func(t *testing.T) {
		logger := &papertrail.LoggerImpl{}
		logger.Infof("Starting %s - test run. . .\n", t.Name())
		var count int64

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt64(&count, 1)
		}))
		defer ts.Close()

		c := NewClient("some string", logger)
		c.baseURL, _ = url.Parse(ts.URL)
		req, err := c.NewRequest("POST", "measurements", &Measurement{})
		if req == nil || err != nil {
			t.Errorf("There was an error creating request")
		}
		resp, err := c.Do(req, nil)
		if err != nil {
			t.Errorf("There was an error: %v", err)
		}
		if resp == nil || resp.StatusCode != http.StatusOK {
			t.Error("Unexpected response")
		}

		if count != 1 {
			t.Errorf("Received requests dont match expected.")
		}

		logger.Infof("Finishing %s - test run. . .\n", t.Name())
	})

	t.Run("Remote error", func(t *testing.T) {
		logger := &papertrail.LoggerImpl{}
		logger.Infof("Starting %s - test run. . .\n", t.Name())
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "I dont like your message.", http.StatusInternalServerError)
		}))
		defer ts.Close()

		c := NewClient("some string", logger)
		c.baseURL, _ = url.Parse(ts.URL)
		req, err := c.NewRequest("POST", "measurements", &Measurement{})
		if req == nil || err != nil {
			t.Errorf("There was an error creating request")
		}
		resp, err := c.Do(req, nil)
		if err == nil {
			t.Errorf("An error is expected")
		}
		if resp != nil && resp.StatusCode == http.StatusOK {
			t.Error("Unexpected response")
		}

		logger.Infof("Finishing %s - test run. . .\n", t.Name())
	})
}
