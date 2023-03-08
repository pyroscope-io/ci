package exec_test

import (
	"io"
	"os"
	"testing"

	"github.com/pyroscope-io/ci/internal/exec"
	"github.com/sirupsen/logrus"
)

func TestCapturingCmdError(t *testing.T) {
	noopLogger := logrus.New()
	noopLogger.SetOutput(io.Discard)
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Running an unknown command, we should get back an error
	cfg := exec.ExecCfg{UploadToCloud: true, Export: true, OutputDir: dir}
	cmdError, err := exec.Exec(noopLogger, []string{"unknown-command"}, cfg)
	if err != nil {
		t.Error("did not expect error to happen")
	}

	if cmdError == nil {
		t.Error("expected cmdError to have happened")
	}

	// Running a valid
	// TODO: what command would be valid cross os?
	cmdError, err = exec.Exec(noopLogger, []string{"echo"}, cfg)
	if err != nil {
		t.Error("did not expect error to happen")
	}

	if cmdError != nil {
		t.Error("expected cmdError to NOT have happened")
	}
}
