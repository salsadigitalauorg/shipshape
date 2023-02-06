package command_test

import (
	"errors"
	"io/fs"
	"os/exec"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/stretchr/testify/assert"
)

func myFuncThatUsesExecCmd() ([]byte, error) {
	cmd := command.ShellCommander("git", "rev-parse", "--abbrev-ref", "HEAD")
	return cmd.Output()
}

func TestExecReplacement(t *testing.T) {
	assert := assert.New(t)

	t.Run("differentStruct", func(t *testing.T) {
		cmd := command.ShellCommander("foo", "bar")
		assert.IsType(command.ExecShellCommand{}, cmd)

		curShellCommander := command.ShellCommander
		defer func() { command.ShellCommander = curShellCommander }()
		command.ShellCommander = func(name string, arg ...string) command.IShellCommand {
			return internal.TestShellCommand{
				OutputterFunc: func() ([]byte, error) {
					return nil, nil
				},
			}
		}
		cmd2 := command.ShellCommander("foo", "bar")
		assert.IsType(internal.TestShellCommand{}, cmd2)
	})

	// temporarily swap the shell commander
	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	t.Run("noError", func(t *testing.T) {
		command.ShellCommander = func(name string, arg ...string) command.IShellCommand {
			return internal.TestShellCommand{
				OutputterFunc: func() ([]byte, error) {
					return []byte("foo"), nil
				},
			}
		}

		out, err := myFuncThatUsesExecCmd()
		assert.Equal([]byte("foo"), out)
		assert.NoError(err)
	})

	t.Run("error", func(t *testing.T) {
		command.ShellCommander = func(name string, arg ...string) command.IShellCommand {
			return internal.TestShellCommand{
				OutputterFunc: func() ([]byte, error) {
					return []byte("foo"), errors.New("bar")
				},
			}
		}

		out, err := myFuncThatUsesExecCmd()
		assert.Equal([]byte("foo"), out)
		assert.Error(err, "bar")
	})
}

func TestGetMsgFromCommandError(t *testing.T) {
	assert := assert.New(t)

	t.Run("pathError", func(t *testing.T) {
		msg := command.GetMsgFromCommandError(&fs.PathError{
			Path: "/path/to/bin",
			Err:  errors.New("foo error"),
		})
		assert.Equal("/path/to/bin: foo error", msg)
	})

	t.Run("exitError", func(t *testing.T) {
		msg := command.GetMsgFromCommandError(&exec.ExitError{
			Stderr: []byte("some error"),
		})
		assert.Equal("some error", msg)
	})

	t.Run("error", func(t *testing.T) {
		msg := command.GetMsgFromCommandError(errors.New("basic error"))
		assert.Equal("basic error", msg)
	})
}
