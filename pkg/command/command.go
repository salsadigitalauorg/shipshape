// Package command provides an interface and implementations for shell commands
// which allow for easy testing and mocking.
//
// It follows the instructions at https://stackoverflow.com/a/74671137/351590
// and https://github.com/schollii/go-test-mock-exec-command which makes use
// of polymorphism to achieve proper testing and mocking.
package command

import (
	"errors"
	"io/fs"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// IShellCommand is an interface for running shell commands.
type IShellCommand interface {
	Output() ([]byte, error)
}

// ExecShellCommand implements IShellCommand.
type ExecShellCommand struct {
	*exec.Cmd
}

// NewExecShellCommander returns a command instance.
func NewExecShellCommander(name string, arg ...string) IShellCommand {
	execCmd := exec.Command(name, arg...)
	return &ExecShellCommand{Cmd: execCmd}
}

func (c *ExecShellCommand) Output() ([]byte, error) {
	log.WithField("command", c).Debug("running command")
	return c.Cmd.Output()
}

// ShellCommander provides a wrapper around the commander to allow for better
// testing and mocking.
var ShellCommander = NewExecShellCommander

// GetMsgFromCommandError attempts to extract the error message from a command
// run's stderr.
func GetMsgFromCommandError(err error) string {
	var pathErr *fs.PathError
	var exitErr *exec.ExitError
	var errMsg string
	if errors.As(err, &pathErr) {
		errMsg = pathErr.Path + ": " + pathErr.Err.Error()
	} else if errors.As(err, &exitErr) {
		errMsg = string(exitErr.Stderr)
	} else {
		errMsg = err.Error()
	}
	return errMsg
}
