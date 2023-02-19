//go:build e2e
// +build e2e

package main_test

import (
	"os"
	"testing"

	"github.com/pyroscope-io/ci/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

// Make the pyroscope-ci binary available to the testscripts
func TestMain(m *testing.M) {
	exitCode := testscript.RunMain(m, map[string]func() int{
		"pyroscope-ci": func() int {
			err := cmd.RootCmd()
			// https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript#RunMain
			if err != nil {
				return 1
			}

			return 0
		},
	})

	os.Exit(exitCode)
}

func TestNodeJest(t *testing.T) {
	containerName, cleanupProxy := StartProxy(t)
	t.Cleanup(cleanupProxy)

	testscript.Run(t, testscript.Params{
		Setup: Setup(
			SetProxyAddressEnvVar(containerName),
			BuildImage("./examples/nodejs/jest", "example-nodejs"),
		),
		Dir: "./examples/nodejs/jest",
	})
}

func TestNodeMocha(t *testing.T) {
	containerName, cleanupProxy := StartProxy(t)
	t.Cleanup(cleanupProxy)

	testscript.Run(t, testscript.Params{
		Setup: Setup(
			SetProxyAddressEnvVar(containerName),
			BuildImage("./examples/nodejs/mocha", "example-mocha"),
		),
		Dir: "./examples/nodejs/mocha",
	})
}

func TestGo(t *testing.T) {
	containerName, cleanupProxy := StartProxy(t)
	t.Cleanup(cleanupProxy)

	testscript.Run(t, testscript.Params{
		Setup: Setup(
			CopyFilesToCwd("./examples/go"),
			SetProxyAddressEnvVar(containerName),
			BuildImage("./examples/go", "example-go"),
		),
		Dir: "./examples/go",
	})
}
