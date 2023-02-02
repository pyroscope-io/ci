package exec_test

import (
	"os"
	"testing"

	"github.com/pyroscope-io/ci/internal/exec"
)

func TestCapturingCmdError(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Running an unknown command, we should get back an error
	cfg := exec.ExecCfg{NoUpload: true, Export: true, OutputDir: dir}
	cmdError, err := exec.Exec([]string{"unknown command"}, cfg)
	if err != nil {
		t.Error("did not expect error to happen")
	}

	if cmdError == nil {
		t.Error("expected cmdError to have happened")
	}

	// Running a valid
	// TODO: what command would be valid cross os?
	cmdError, err = exec.Exec([]string{"echo"}, cfg)
	if err != nil {
		t.Error("did not expect error to happen")
	}

	if cmdError != nil {
		t.Error("expected cmdError to NOT have happened")
	}

}
