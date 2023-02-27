package flamegraphdotcom_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pyroscope-io/ci/internal/upload/flamegraphdotcom"
	"github.com/sirupsen/logrus"
)

func TestUpload(t *testing.T) {
	noopLogger := logrus.New()
	noopLogger.SetOutput(io.Discard)

	expected := []flamegraphdotcom.Response{{Url: "my-url"}}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: handle multiple
		if err := json.NewEncoder(w).Encode(expected[0]); err != nil {
			panic(err)
		}
	}))
	defer svr.Close()

	uploader := flamegraphdotcom.NewUploader(noopLogger, svr.URL)

	response, err := uploader.UploadMultiple(context.TODO(), []string{"./testadata/single.json"})
	if err != nil {
		t.Fatalf("expected err to be nil got %v", err)
	}

	if len(expected) != len(response) {
		t.Fatalf("expected to find %d responses but got %d", len(expected), len(response))
	}

	for i := range response {
		if response[i].Url != expected[i].Url {
			t.Fatalf("expected response url to be '%s' but got '%s'", expected[i].Url, response[i].Url)
		}
	}
}
