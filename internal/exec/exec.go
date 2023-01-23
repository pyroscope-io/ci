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

// Exec executes a program and exports its captured profiles
// It works by creating an in-memory pyroscope server
// Then overwriting the default serverAddress using the PYROSCOPE_ADHOC_SERVER_ADDRESS env var
// Which then is either uploaded to a) a pyroscope server that supports the /ci API
// Or b) to a local directory
//
// Notice that it returns 2 different errors:
// cmdError refers to the error of the command exec'd
// and err to any other error
func Exec(args []string, cfg ExecCfg) (cmdError error, err error) {
	logger := logrus.StandardLogger()
	lvl, _ := logrus.ParseLevel(cfg.LogLevel)
	logger.SetLevel(lvl)

	runner := NewRunner(logger)

	logger.Debug("exec'ing command")
	ingestedItems, duration, cmdError := runner.Run(args)
	if err != nil {
		logger.Errorf("process errored: %s. will still try to upload ingested data", err)
		//return err
	}

	if len(ingestedItems) <= 0 {
		logger.Info("No profiles were ingested. Nothing to export")
		return cmdError, err
	}

	if cfg.NoUpload {
		logger.Debug("exporting files since NoUpload flag is on")
		exporter := NewExporter(logger, cfg.OutputDir)
		return cmdError, exporter.Export(ingestedItems)
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
	return cmdError, uploader.Upload(context.Background(), ingestedItems)
}
