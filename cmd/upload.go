package cmd

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pyroscope-io/ci/internal/upload"
	"github.com/sirupsen/logrus"
)

func uploadCmd() *ffcli.Command {
	uploadFlagSet := flag.NewFlagSet("upload", flag.ExitOnError)
	commitSHA := uploadFlagSet.String("commitSHA", "", "the commit sha")
	applicationName := uploadFlagSet.String("applicationName", "", "")
	branch := uploadFlagSet.String("branch", "", "")
	serverAddress := uploadFlagSet.String("serverAddress", "http://localhost:4040", "")
	apiKey := uploadFlagSet.String("apiKey", "", "")
	spyName := uploadFlagSet.String("spyName", "", "")
	// TODO: unit
	// unit := uploadFlagSet.String("unit", "", "")
	//	date := uploadFlagSet.String("date", "", "")

	return &ffcli.Command{
		Name:       "upload",
		ShortUsage: "",
		ShortHelp:  "",
		FlagSet:    uploadFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("at least a file needs to be specified")
			}

			if len(args) > 1 && *applicationName != "" {
				return fmt.Errorf("multiple files are passed but the same applicationName would apply. " +
					"this is most likely not wanted. if that's the desired behaviour, run the 'upload' command once per file. aborting")
			}

			logger := logrus.New()
			uploader := upload.NewUploader(logger)

			now := time.Now()
			return uploader.UploadMultiple(ctx, upload.UploadMultipleCfg{
				APIKey:        *apiKey,
				AppName:       *applicationName,
				Branch:        *branch,
				Date:          now,
				CommitSHA:     *commitSHA,
				Filepath:      args,
				ServerAddress: *serverAddress,
				SpyName:       *spyName,
			})
		},
	}
}
