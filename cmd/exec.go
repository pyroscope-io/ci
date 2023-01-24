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
	outputDir := execFlagSet.String("outputDir", "pyroscope-ci", "where the generated profiles will be saved to. only available if --no-upload is set")
	serverAddress := execFlagSet.String("serverAddress", "https://pyroscope.cloud", "")
	apiKey := execFlagSet.String("apiKey", "", "")
	commitSHA := execFlagSet.String("commitSHA", "", "the commit sha")
	branch := execFlagSet.String("branch", "", "")
	noUpload := execFlagSet.Bool("noUpload", false, "whether to upload automatically or to store into a local directory")
	logLevel := execFlagSet.String("logLevel", "info", "")

	return &ffcli.Command{
		Name:       "exec",
		ShortHelp:  "exec a command and save its profiling data",
		ShortUsage: "pyro-ci exec -- make test",
		FlagSet:    execFlagSet,
		Exec: func(_ context.Context, args []string) error {
			if len(args) <= 0 {
				return fmt.Errorf("at least one argument is required")
			}

			cmdError, err := exec.Exec(args, exec.ExecCfg{
				OutputDir:     *outputDir,
				APIKey:        *apiKey,
				ServerAddress: *serverAddress,
				CommitSHA:     *commitSHA,
				NoUpload:      *noUpload,
				Branch:        *branch,
				LogLevel:      *logLevel,
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
