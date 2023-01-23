package cmd

import (
	"context"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func RootCmd() error {
	root := &ffcli.Command{
		Subcommands: []*ffcli.Command{installCmd(), execCmd()},
	}

	return root.ParseAndRun(context.Background(), os.Args[1:])
}
