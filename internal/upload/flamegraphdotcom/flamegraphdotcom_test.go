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
	"github.com/sirupsen/logrus"
)

func TestUpload(t *testing.T) {
	noopLogger := logrus.New()
	noopLogger.SetOutput(io.Discard)

	expected := []flamegraphdotcom.UploadedFlamegraph{
		{Url: "my-url", Filename: "./testdata/single.json"},
		{Url: "my-url", Filename: "./testdata/inuse_objects.json"},
	}
	findFromFilename := func(filename string) flamegraphdotcom.UploadedFlamegraph {
		for _, exp := range expected {
			if exp.Filename == filename {
				return exp
			}
		}

		panic(fmt.Sprintf("could not find file_name %s", filename))
	}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get("file_name")
		exp := findFromFilename(filename)
		if err := json.NewEncoder(w).Encode(flamegraphdotcom.FlamegraphDotComResponse{
			Url: exp.Url,
		}); err != nil {
			panic(err)
		}
	}))
	defer svr.Close()

	uploader := flamegraphdotcom.NewUploader(noopLogger, svr.URL)

	filenames := make([]string, len(expected))
	for i, f := range expected {
		filenames[i] = f.Filename
	}

	response, err := uploader.UploadMultiple(context.TODO(), filenames)
	if err != nil {
		t.Fatalf("expected err to be nil got %v", err)
	}

	if len(expected) != len(response) {
		t.Fatalf("expected to find %d responses but got %d", len(expected), len(response))
	}

	for i, resp := range response {
		exp := findFromFilename(resp.Filename)

		if resp.Url != exp.Url {
			t.Fatalf("expected response url to be '%s' but got '%s'", expected[i].Url, response[i].Url)
		}

		if resp.Filename != exp.Filename {
			t.Fatalf("expected response filename to be '%s' but got '%s'", expected[i].Filename, response[i].Filename)
		}
	}
}
