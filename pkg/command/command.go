// Package command provides an interface and implementations for shell commands
// which allow for easy testing and mocking.
//
// It follows the instructions at https://stackoverflow.com/a/74671137/351590
// and https://github.com/schollii/go-test-mock-exec-command which makes use
// of polymorphism to achieve proper testing and mocking.
package command

import "os/exec"

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
	return ExecShellCommand{Cmd: execCmd}
}

// ShellCommander provides a wrapper around the commander to allow for better
// testing and mocking.
var ShellCommander = NewExecShellCommander
