package exec

import (
	"context"
	"fmt"

	"github.com/pyroscope-io/ci/internal/upload/flamegraphdotcom"
	"github.com/pyroscope-io/ci/internal/upload/pyroscopecloud"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
)

type ExecCfg struct {
	OutputDir         string
	ServerAddress     string
	APIKey            string
	CommitSHA         string
	Branch            string
	NoUpload          bool
	Export            bool
	UploadToPublicApi bool
	LogLevel          string
}

// Exec executes a program and exports its captured profiles
// It works by creating an in-memory pyroscope server
// Then overwriting the default serverAddress using the PYROSCOPE_ADHOC_SERVER_ADDRESS env var
// Which then is uploaded to
// a) a pyroscope server that supports the /ci API
// b) to a local directory
// c) to a public API (here flamegraph.com)
// d) any combination of above options
//
// Notice that it returns 2 different errors:
// cmdError refers to the error of the command exec'd
// and err to any other error
func Exec(args []string, cfg ExecCfg) (cmdError error, err error) {
	logger := logrus.StandardLogger()
	lvl, _ := logrus.ParseLevel(cfg.LogLevel)
	logger.SetLevel(lvl)

	runner := NewRunner(logger)

	if !cfg.Export && cfg.NoUpload && !cfg.UploadToPublicApi {
		logger.Warn("not uploading, exporting and not uploading to public api, this does not look intended")
		return nil, nil
	}

	logger.Debug("exec'ing command")
	ingestedItems, duration, cmdError := runner.Run(args)
	if err != nil {
		logger.Errorf("process errored: %s. will still try to upload ingested data", err)
	}

	if len(ingestedItems) <= 0 {
		logger.Info("No profiles were ingested. Nothing to export")
		return cmdError, err
	}

	if cfg.Export {
		logger.Debug("exporting files to ", cfg.OutputDir)
		exporter := NewExporter(logger, cfg.OutputDir)
		err = exporter.Export(ingestedItems)
		if err != nil {
			return cmdError, fmt.Errorf("error exporting data: %w", err)
		}
	}

	if cfg.NoUpload {
		logger.Debug("not uploading to the cloud since --noUpload flag is turned on")
	} else {
		ciCtx := DetectContext(cfg)
		uploader := pyroscopecloud.NewUploader(logger, pyroscopecloud.UploadConfig{
			// Generate a shared ID that will group different profiles
			Id:            ksuid.New(),
			ApiKey:        cfg.APIKey,
			ServerAddress: cfg.ServerAddress,
			CommitSHA:     ciCtx.CommitSHA,
			Branch:        ciCtx.BranchName,
			Duration:      duration,
		})

		logger.Debugf("uploading %d profile(s) to cloud", len(ingestedItems))
		err = uploader.Upload(context.Background(), ingestedItems)
		if err != nil {
			return cmdError, fmt.Errorf("uploading profiles: %w", err)
		}
	}

	if cfg.UploadToPublicApi {
		logger.Debugf("uploading %d profile(s)", len(ingestedItems))
		flamegraphUploader := flamegraphdotcom.NewUploader(logger, "")
		res, err := flamegraphUploader.Upload(context.Background(), ingestedItems)
		if err != nil {
			return cmdError, fmt.Errorf("uploading profiles to public api: %w", err)
		}

		logger.Info("Profiles have been uploaded to a public API:")
		for _, r := range res {
			logger.Info("\t", r.AppName, "\t", r.Url)
		}
	}

	return cmdError, nil
}
