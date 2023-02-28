package flamegraphdotcom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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
	Url      string
	Filename string
}

func (u *Uploader) UploadMultiple(ctx context.Context, filepath []string) ([]UploadedFlamegraph, error) {
	g, _ := errgroup.WithContext(ctx)

	responses := make([]UploadedFlamegraph, len(filepath))

	for i, f := range filepath {
		f := f
		i := i

		g.Go(func() error {
			u.logger.Debug("uploading", f)
			o, err := u.uploadSingle(ctx, f)
			responses[i] = UploadedFlamegraph{
				Url:      o.Url,
				Filename: f,
			}

			return err
		})
	}

	return responses, g.Wait()
}

type FlamegraphDotComResponse struct {
	Url string `json:"url"`
}

func (u *Uploader) uploadSingle(ctx context.Context, filepath string) (FlamegraphDotComResponse, error) {
	var response FlamegraphDotComResponse

	file, err := os.Open(filepath)
	if err != nil {
		return response, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/upload/v1", u.serverURL), file)
	if err != nil {
		return response, err
	}
	q := req.URL.Query()
	q.Add("file_name", filepath)
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
