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

func (u *Uploader) UploadMultiple(ctx context.Context, filepath []string) ([]Response, error) {
	g, _ := errgroup.WithContext(ctx)

	responses := make([]Response, len(filepath))

	for i, f := range filepath {
		f := f
		i := i

		g.Go(func() error {
			u.logger.Debug("uploading", f)
			o, err := u.uploadSingle(ctx, f)
			responses[i] = o

			return err
		})
	}

	return responses, g.Wait()
}

type Response struct {
	Url string `json:"url"`
}

func (u *Uploader) uploadSingle(ctx context.Context, filepath string) (Response, error) {
	var response Response

	file, err := os.Open(filepath)
	if err != nil {
		return response, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/upload/v1", u.serverURL), file)
	if err != nil {
		return response, err
	}

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
