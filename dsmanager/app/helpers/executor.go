package helpers

import (
	"bytes"
	"os/exec"

	"golang.org/x/xerrors"
)

// Executor is an interface that alllows one to execute commands. This
// abstraction allows us to mock a traditional "os/exec" command in the tests.
type Executor interface {
	Run(args ...string) (bytes.Buffer, error)
}

// OSExecutor implements the Executor interface with os/exec
type OSExecutor struct {
}

// NewOSExecutor returns a new OS Executor
func NewOSExecutor() OSExecutor {
	return OSExecutor{}
}

// Run executes the command given in args. The first arg must be the name of the
// command
func (ose OSExecutor) Run(args ...string) (bytes.Buffer, error) {
	var outb, errb bytes.Buffer

	if len(args) == 0 {
		return outb, xerrors.New("args should have at least one element")
	}

	var cmd *exec.Cmd
	if len(args) == 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		return outb, xerrors.Errorf("failed to execute command: %s - Output: %s - Err: %s",
			err.Error(), outb.String(), errb.String())
	}

	return outb, nil
}
