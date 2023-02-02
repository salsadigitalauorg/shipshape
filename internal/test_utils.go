package internal

import (
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
)

type TestShellCommand struct {
	OutputterFunc func() ([]byte, error)
}

func (sc TestShellCommand) Output() ([]byte, error) {
	return sc.OutputterFunc()
}

// DumbShellCommander simply returns nil - can be used to simply override
// commands where we don't expect anything from the output or err.
var DumbShellCommander = func(name string, arg ...string) command.IShellCommand {
	return TestShellCommand{
		OutputterFunc: func() ([]byte, error) {
			return nil, nil
		},
	}
}

// ShellCommanderMaker is a commander generator that can return the provided
// stdout or stderr, and can also update a given variable with the generated
// command.
func ShellCommanderMaker(out *string, err error, updateVar *string) func(name string, arg ...string) command.IShellCommand {
	return func(name string, arg ...string) command.IShellCommand {
		if updateVar != nil {
			fullCmd := append([]string{name}, arg...)
			*updateVar = strings.Join(fullCmd, " ")
		}
		var stdout []byte
		if out != nil {
			stdout = []byte(*out)
		}
		return TestShellCommand{
			OutputterFunc: func() ([]byte, error) {
				return stdout, err
			},
		}
	}
}
