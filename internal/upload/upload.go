package upload

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Uploader struct {
	logger     *logrus.Logger
	httpClient *http.Client
}

func NewUploader(logger *logrus.Logger) *Uploader {
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	return &Uploader{
		logger:     logger,
		httpClient: httpClient,
	}
}

type UploadMultipleCfg struct {
	AppName       string
	Branch        string
	Date          time.Time
	CommitSHA     string
	Filepath      []string
	ServerAddress string
	APIKey        string
	SpyName       string
}

// Upload uploads files to the target server's /api/ci-events endpoint
// It assumes all cfg values are non-zero
func (u *Uploader) UploadMultiple(ctx context.Context, cfg UploadMultipleCfg) error {
	g, _ := errgroup.WithContext(ctx)

	for _, f := range cfg.Filepath {
		f := f

		g.Go(func() error {
			singleCfg := UploadSingleCfg{
				appName:       cfg.AppName,
				branch:        cfg.Branch,
				commitSHA:     cfg.CommitSHA,
				serverAddress: cfg.ServerAddress,
				spyName:       cfg.SpyName,
				apiKey:        cfg.APIKey,
				date:          cfg.Date,
				filepath:      f,
			}

			u.logger.Debug("uploading", singleCfg)
			return u.upload(singleCfg)
		})
	}

	return g.Wait()
}

type UploadSingleCfg struct {
	appName       string
	branch        string
	date          time.Time
	commitSHA     string
	filepath      string
	serverAddress string
	apiKey        string
	spyName       string
}

// TODO
func (u *Uploader) upload(cfg UploadSingleCfg) error {
	file, err := os.Open(cfg.filepath)
	if err != nil {
		return err
	}

	// TODO: timeouts and whatnot
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/ci-events", cfg.serverAddress), file)
	if err != nil {
		return err
	}

	if cfg.apiKey != "" {
		req.Header.Add("Authorization", "Bearer "+cfg.apiKey)
	}

	q := req.URL.Query()
	q.Add("date", cfg.date.Format(time.RFC3339))
	q.Add("name", cfg.appName)
	q.Add("branch", cfg.branch)
	q.Add("commitSHA", cfg.commitSHA)
	q.Add("spyName", cfg.spyName)

	req.URL.RawQuery = q.Encode()

	res, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// read all the response body
	var respBody []byte
	_, err = res.Body.Read(respBody)
	if err != nil {
		return fmt.Errorf("read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("statusCode '%d'. body '%s'", res.StatusCode, respBody)
	}

	return nil
}

//func (u *Uploader) upload(appName string, branch string, date time.Time, commitSHA string, filepath string, serverAddress string, apiKey string, spyName string) error {
//	file, err := os.Open(filepath)
//	if err != nil {
//		return err
//	}
//
//	// TODO: timeouts and whatnot
//	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/ci-events", serverAddress), file)
//	if err != nil {
//		return err
//	}
//
//	if apiKey != "" {
//		req.Header.Add("Authorization", "Bearer "+apiKey)
//	}
//
//	q := req.URL.Query()
//	q.Add("date", date.Format(time.RFC3339))
//	q.Add("name", appName)
//	q.Add("branch", branch)
//	q.Add("commitSHA", commitSHA)
//	q.Add("spyName", spyName)
//
//	req.URL.RawQuery = q.Encode()
//
//	fmt.Println("doing")
//	res, err := u.httpClient.Do(req)
//	if err != nil {
//		return err
//	}
//	defer res.Body.Close()
//
//	// read all the response body
//	respBody, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		return fmt.Errorf("read response body: %v", err)
//	}
//
//	if res.StatusCode != http.StatusOK {
//		return fmt.Errorf("statusCode '%d'. body '%s'", res.StatusCode, respBody)
//	}
//
//	return nil
//}
