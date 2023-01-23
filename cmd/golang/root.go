package golang

import (
	"context"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func GoCmd() *ffcli.Command {
	cmd := &ffcli.Command{
		Name:        "go",
		ShortHelp:   "subcommands scoped for go",
		Subcommands: []*ffcli.Command{installCmd()},
	}
	cmd.Exec = func(ctx context.Context, args []string) error {
		fmt.Println(ffcli.DefaultUsageFunc(cmd))
		return nil
	}

	return cmd
}
