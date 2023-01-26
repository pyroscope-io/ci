package golang

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pyroscope-io/ci/internal/golang/install"
)

func installCmd() *ffcli.Command {
	installFlagSet := flag.NewFlagSet("install", flag.ExitOnError)
	profileTypesFlag := installFlagSet.String("profileTypes", "all",
		fmt.Sprintf("list of profileTypes, separated by comma. available types are: %s", install.AvailableProfileTypes))
	appName := installFlagSet.String("applicationName", "", "the name of the application")

	cmd := &ffcli.Command{
		Name:       "install",
		ShortUsage: "pyroscope-ci go install {packagePath}",
		ShortHelp:  "Installs the pyroscope agent into test packages",
		LongHelp: "Given a (list of) {packagePath}, it will recursively find all packages that contains tests." +
			"Then it will generate a `pyroscope_test.go` file for each package, using the configuration passed as cli flags from this command.",
		FlagSet: installFlagSet,
		Exec: func(_ context.Context, args []string) error {
			if len(args) <= 0 {
				return fmt.Errorf("at least one path needs to be specified")
			}

			profileTypes, err := install.ProfileTypesToCode(*profileTypesFlag)
			if err != nil {
				return err
			}

			if *appName == "" {
				return fmt.Errorf("--applicationName is required")
			}

			for _, a := range args {
				if install.Install(a, *appName, profileTypes); err != nil {
					return err
				}
			}

			fmt.Println("installed.")
			return nil
		},
	}

	return cmd
}
