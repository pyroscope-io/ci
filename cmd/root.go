package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func RootCmd() {
	root := &ffcli.Command{
		Subcommands: []*ffcli.Command{installCmd(), execCmd(), uploadCmd()},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
