package flamegraphdotcom_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pyroscope-io/ci/internal/upload/flamegraphdotcom"
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
	"github.com/sirupsen/logrus"
)

func TestUpload(t *testing.T) {
	noopLogger := logrus.New()
	noopLogger.SetOutput(io.Discard)

	data := map[string]flamebearer.FlamebearerProfile{
		"my-app1": {},
		"my-app2": {},
	}

	expected := []flamegraphdotcom.UploadedFlamegraph{
		{URL: "my-url-for-app1", AppName: "my-app1"},
		{URL: "my-url-for-app2", AppName: "my-app2"},
	}
	findFromAppName := func(appName string) flamegraphdotcom.UploadedFlamegraph {
		for _, exp := range expected {
			if exp.AppName == appName {
				return exp
			}
		}

		panic(fmt.Sprintf("could not find file_name '%s'", appName))
	}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get("file_name")
		exp := findFromAppName(filename)
		if err := json.NewEncoder(w).Encode(flamegraphdotcom.FlamegraphDotComResponse{
			URL: exp.URL,
		}); err != nil {
			panic(err)
		}
	}))
	defer svr.Close()

	uploader := flamegraphdotcom.NewUploader(noopLogger, svr.URL)

	filenames := make([]string, len(expected))
	for i, f := range expected {
		filenames[i] = f.AppName
	}

	response, err := uploader.Upload(context.TODO(), data)
	if err != nil {
		t.Fatalf("expected err to be nil got %v", err)
	}

	if len(expected) != len(response) {
		t.Fatalf("expected to find %d responses but got %d", len(expected), len(response))
	}

	for i, resp := range response {
		exp := findFromAppName(resp.AppName)

		if resp.URL != exp.URL {
			t.Fatalf("expected response url to be '%s' but got '%s'", expected[i].URL, response[i].URL)
		}

		if resp.AppName != exp.AppName {
			t.Fatalf("expected response filename to be '%s' but got '%s'", expected[i].AppName, response[i].AppName)
		}
	}
}
