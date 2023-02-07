package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type UploadConfig struct {
	apiKey        string
	serverAddress string
	commitSHA     string
	branch        string
	duration      time.Duration
	id            ksuid.KSUID
}

type Uploader struct {
	logger     *logrus.Logger
	httpClient *http.Client
	cfg        UploadConfig
}

func NewUploader(logger *logrus.Logger, cfg UploadConfig) *Uploader {
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	return &Uploader{
		logger:     logger,
		httpClient: httpClient,
		cfg:        cfg,
	}
}

// Upload uploads files to the target server's /api/ci-events endpoint
// It assumes all cfg values are non-zero
func (u *Uploader) Upload(ctx context.Context, items map[string]flamebearer.FlamebearerProfile) error {
	g, ctx := errgroup.WithContext(ctx)

	for name, f := range items {
		f := f
		name := name

		g.Go(func() error {
			u.logger.Debug("uploading ", name)
			return u.uploadSingle(ctx, f)
		})
	}

	return g.Wait()
}

func (u *Uploader) uploadSingle(_ context.Context, item flamebearer.FlamebearerProfile) error {
	serverAddress := u.cfg.serverAddress
	apiKey := u.cfg.apiKey
	commitSHA := u.cfg.commitSHA
	branch := u.cfg.branch
	duration := u.cfg.duration.String()
	id := u.cfg.id.String()

	date := time.Now()

	marshalled, err := json.Marshal(item)
	if err != nil {
		return err
	}

	file := bytes.NewReader(marshalled)

	// TODO: get the whole url from the config?
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/ci-events", serverAddress), file)
	if err != nil {
		return err
	}

	if apiKey != "" {
		req.Header.Add("Authorization", "Bearer "+apiKey)
	}

	q := req.URL.Query()
	q.Add("date", date.Format(time.RFC3339))
	q.Add("commitSHA", commitSHA)
	q.Add("branch", branch)
	q.Add("duration", duration)
	q.Add("id", id)

	req.URL.RawQuery = q.Encode()

	res, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)

	if err != nil {
		return fmt.Errorf("error reading response body: '%v'. the request had statusCode %d", err, res.StatusCode)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("statusCode '%d'. body '%s'", res.StatusCode, respBody)
	}

	return nil
}
