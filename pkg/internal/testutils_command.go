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

// ShellCommanderMaker is a commander generator that can return the provided
// stdout or stderr, and can also update a given variable with the generated
// command.
func ShellCommanderMaker(out *string, err error, generatedCommand *string) func(name string, arg ...string) command.IShellCommand {
	return func(name string, arg ...string) command.IShellCommand {
		if generatedCommand != nil {
			fullCmd := name
			for _, a := range arg {
				// Add quotes when there are spaces.
				if len(strings.Fields(a)) > 1 {
					a = "'" + a + "'"
				}
				fullCmd += " " + a
			}
			*generatedCommand = fullCmd
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
