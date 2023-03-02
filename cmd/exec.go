package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pyroscope-io/ci/internal/exec"
	"github.com/sirupsen/logrus"
)

func execCmd() *ffcli.Command {
	execFlagSet := flag.NewFlagSet("exec", flag.ExitOnError)

	exportLocally := execFlagSet.Bool("exportLocally", false, "exports data to a local directory, used in conjunction with --outputDir")
	outputDir := execFlagSet.String("outputDir", "pyroscope-ci-output", "where the generated profiles will be saved to. only available if --exportLocally is set")
	cloudServerAddress := execFlagSet.String("serverAddress", "https://pyroscope.cloud", "")
	apiKey := execFlagSet.String("apiKey", "", "")
	commitSHA := execFlagSet.String("commitSHA", "", "the commit sha")
	branch := execFlagSet.String("branch", "", "")
	uploadToCloud := execFlagSet.Bool("uploadToCloud", true, "uploads to pyroscope cloud")
	uploadToPublicAPI := execFlagSet.Bool("uploadToPublicAPI", false, "uploads to public API (flamegraph.com)")
	logLevel := execFlagSet.String("logLevel", "info", "")

	return &ffcli.Command{
		Name:       "exec",
		ShortHelp:  "exec a command and save its profiling data",
		ShortUsage: "pyroscope-ci exec -- make test",
		FlagSet:    execFlagSet,
		Exec: func(_ context.Context, args []string) error {
			if len(args) <= 0 {
				return fmt.Errorf("at least one argument is required")
			}

			cmdError, err := exec.Exec(args, exec.ExecCfg{
				OutputDir:         *outputDir,
				APIKey:            *apiKey,
				ServerAddress:     *cloudServerAddress,
				CommitSHA:         *commitSHA,
				NoUpload:          *uploadToCloud,
				Branch:            *branch,
				LogLevel:          *logLevel,
				Export:            *exportLocally,
				UploadToPublicAPI: *uploadToPublicAPI,
			})

			// If exec failed, print it first
			if err != nil {
				return err
			}

			// Add additional context, since it may be weird to just see "exit eror code X"
			if cmdError != nil {
				logrus.Error("error in spawned command")
			}
			return cmdError
		},
	}
}
