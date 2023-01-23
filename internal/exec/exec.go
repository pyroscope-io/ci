package exec

import (
	"context"

	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
)

type ExecCfg struct {
	OutputDir     string
	ServerAddress string
	APIKey        string
	CommitSHA     string
	Branch        string
	NoUpload      bool
	LogLevel      string
}

func Exec(args []string, cfg ExecCfg) error {
	logger := logrus.StandardLogger()
	lvl, _ := logrus.ParseLevel(cfg.LogLevel)
	logger.SetLevel(lvl)

	runner, err := NewRunner(logger)
	if err != nil {
		return err
	}

	logger.Debug("exec'ing command")
	ingestedItems, duration, err := runner.Run(args)
	if err != nil {
		logger.Errorf("process errored: %s. will still try to upload ingested data", err)
		//return err
	}

	if len(ingestedItems) <= 0 {
		logger.Info("No profiles were ingested. Nothing to export")
		return nil
	}

	if cfg.NoUpload {
		logger.Debug("exporting files since NoUpload flag is on")
		exporter := NewExporter(logger, cfg.OutputDir)
		return exporter.Export(ingestedItems)
	}

	ciCtx := DetectContext(cfg)
	uploader := NewUploader(logger, UploadConfig{
		// Generate a shared ID that will group different profiles
		id:            ksuid.New(),
		apiKey:        cfg.APIKey,
		serverAddress: cfg.ServerAddress,
		commitSHA:     ciCtx.CommitSHA,
		branch:        ciCtx.BranchName,
		duration:      duration,
	})

	logger.Debugf("uploading %d profile(s)", len(ingestedItems))
	return uploader.Upload(context.Background(), ingestedItems)
}
