package command_test

import (
	"errors"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/stretchr/testify/assert"
)

type myShellCommand struct {
	OutputterFunc func() ([]byte, error)
}

func (sc myShellCommand) Output() ([]byte, error) {
	return sc.OutputterFunc()
}

func myFuncThatUsesExecCmd() ([]byte, error) {
	cmd := command.ShellCommander("git", "rev-parse", "--abbrev-ref", "HEAD")
	return cmd.Output()
}

func TestExecReplacement(t *testing.T) {
	assert := assert.New(t)

	// temporarily swap the shell commander
	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	t.Run("noError", func(t *testing.T) {
		command.ShellCommander = func(name string, arg ...string) command.IShellCommand {
			return myShellCommand{
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
			return myShellCommand{
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
