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
	UploadToCloud     bool
	Export            bool
	UploadToPublicAPI bool
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
func Exec(logger *logrus.Logger, args []string, cfg ExecCfg) (cmdError error, err error) {
	runner := NewRunner(logger)

	if !cfg.Export && !cfg.UploadToCloud && !cfg.UploadToPublicAPI {
		logger.Warn("not uploading, exporting and not uploading to public api, this does not look intended")
		return nil, nil
	}

	logger.Debug("exec'ing command")
	ingestedItems, duration, cmdError := runner.Run(args)
	if cmdError != nil {
		logger.Debug("process errored. will still try to upload ingested data", cmdError)
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

	if cfg.UploadToCloud {
		ciCtx := DetectContext(cfg)
		uploader := pyroscopecloud.NewUploader(logger, pyroscopecloud.UploadConfig{
			// Generate a shared ID that will group different profiles
			ID:            ksuid.New(),
			APIKey:        cfg.APIKey,
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

	if cfg.UploadToPublicAPI {
		logger.Debugf("uploading %d profile(s)", len(ingestedItems))
		flamegraphUploader := flamegraphdotcom.NewUploader(logger, "")
		res, err := flamegraphUploader.Upload(context.Background(), ingestedItems)
		if err != nil {
			return cmdError, fmt.Errorf("uploading profiles to public api: %w", err)
		}

		logger.Info("Profiles have been uploaded to a public API:")
		for _, r := range res {
			logger.Info("\t", r.AppName, "\t", r.URL)
		}
	}

	return cmdError, nil
}
