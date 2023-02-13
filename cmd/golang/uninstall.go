package golang

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pyroscope-io/ci/internal/golang/install"
)

func uninstallCmd() *ffcli.Command {
	installFlagSet := flag.NewFlagSet("uninstall", flag.ExitOnError)

	cmd := &ffcli.Command{
		Name:       "uninstall",
		ShortUsage: "pyroscope-ci go uninstall {packagePath}",
		ShortHelp:  "Uninstalls the pyroscope agent from test packages",
		FlagSet:    installFlagSet,
		Exec: func(_ context.Context, args []string) error {
			if len(args) <= 0 {
				return fmt.Errorf("at least one path needs to be specified")
			}

			for _, a := range args {
				output, err := install.Uninstall(a)
				if err != nil {
					return err
				}

				for _, v := range output {
					fmt.Println("Deleted", v.Path)
				}
			}

			return nil
		},
	}

	return cmd
}
