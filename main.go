package main

import (
	"fmt"
	"os"

	"github.com/pyroscope-io/ci/cmd"
)

func main() {
	if err := cmd.RootCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
