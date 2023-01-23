package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pyroscope-io/ci/cmd/golang"
)

func RootCmd() error {
	rootCmd := &ffcli.Command{
		Subcommands: []*ffcli.Command{golang.GoCmd(), execCmd()},
	}
	rootCmd.Exec = func(ctx context.Context, args []string) error {
		fmt.Println(ffcli.DefaultUsageFunc(rootCmd))
		return nil
	}

	return rootCmd.ParseAndRun(context.Background(), os.Args[1:])
}
