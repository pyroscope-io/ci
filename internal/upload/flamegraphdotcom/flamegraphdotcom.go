package flamegraphdotcom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Uploader struct {
	logger     *logrus.Logger
	httpClient *http.Client
	serverURL  string
}

func NewUploader(logger *logrus.Logger, serverURL string) *Uploader {
	if serverURL == "" {
		serverURL = "https://flamegraph.com"
	}

	httpClient := &http.Client{
		Timeout: time.Second * 60,
	}

	return &Uploader{
		logger:     logger,
		httpClient: httpClient,
		serverURL:  serverURL,
	}
}

type UploadedFlamegraph struct {
	Url     string
	AppName string
}

func (u *Uploader) Upload(ctx context.Context, items map[string]flamebearer.FlamebearerProfile) ([]UploadedFlamegraph, error) {
	g, _ := errgroup.WithContext(ctx)

	responses := make([]UploadedFlamegraph, 0)

	var mu sync.Mutex
	for appName, f := range items {
		f := f
		appName := appName

		g.Go(func() error {
			u.logger.Debug("uploading ", f.Metadata.Name)
			o, err := u.uploadSingle(ctx, appName, f)
			mu.Lock()
			responses = append(responses, UploadedFlamegraph{
				Url:     o.Url,
				AppName: appName,
			})
			mu.Unlock()

			return err
		})
	}

	return responses, g.Wait()
}

type FlamegraphDotComResponse struct {
	Url string `json:"url"`
}

func (u *Uploader) uploadSingle(ctx context.Context, appName string, item flamebearer.FlamebearerProfile) (FlamegraphDotComResponse, error) {
	var response FlamegraphDotComResponse

	marshalled, err := json.Marshal(item)
	if err != nil {
		return response, err
	}

	file := bytes.NewReader(marshalled)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/upload/v1", u.serverURL), file)
	if err != nil {
		return response, err
	}
	q := req.URL.Query()
	q.Add("file_name", appName)
	req.URL.RawQuery = q.Encode()

	res, err := u.httpClient.Do(req)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	// read all the response body
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return response, fmt.Errorf("read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return response, fmt.Errorf("statusCode '%d'. body '%s'", res.StatusCode, respBody)
	}

	err = json.Unmarshal(respBody, &response)
	return response, err
}
