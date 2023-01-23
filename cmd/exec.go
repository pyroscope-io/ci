package cmd

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pyroscope-io/ci/internal/exec"
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
		ShortUsage: "",
		ShortHelp:  "",
		FlagSet:    execFlagSet,
		Exec: func(_ context.Context, args []string) error {
			cmdError, err := exec.Exec(args, exec.ExecCfg{
				OutputDir:     *outputDir,
				APIKey:        *apiKey,
				ServerAddress: *serverAddress,
				CommitSHA:     *commitSHA,
				NoUpload:      *noUpload,
				Branch:        *branch,
				LogLevel:      *logLevel,
			})
			if err != nil {
				return err
			}
			return cmdError
		},
	}
}
