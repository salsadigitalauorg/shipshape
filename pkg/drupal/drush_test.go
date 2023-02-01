package drupal_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/stretchr/testify/assert"
)

func TestDrushCommandMerge(t *testing.T) {
	assert := assert.New(t)

	dc := drupal.DrushCommand{
		DrushPath: "/path/to/drush",
		Alias:     "alias1",
		Args:      []string{"arg1", "arg2"},
	}

	dc.Merge(drupal.DrushCommand{DrushPath: "/new/path/to/drush"})
	assert.Equal("/new/path/to/drush", dc.DrushPath)

	dc.Merge(drupal.DrushCommand{Alias: "alias2"})
	assert.Equal("alias2", dc.Alias)

	dc.Merge(drupal.DrushCommand{Args: []string{"arg2", "arg3"}})
	assert.ElementsMatch([]string{"arg2", "arg3"}, dc.Args)
}

func TestDrushExec(t *testing.T) {
	assert := assert.New(t)

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	// Command not found.
	t.Run("commandNotFound", func(t *testing.T) {
		command.ShellCommander = func(name string, arg ...string) command.IShellCommand {
			return internal.TestShellCommand{
				OutputterFunc: func() ([]byte, error) {
					return nil, errors.New("bash: drushfoo: command not found")
				},
			}
		}

		_, err := drupal.Drush("", "", []string{"status"}).Exec()
		assert.Error(err, "bash: drushfoo: command not found")
	})

	t.Run("ok", func(t *testing.T) {
		command.ShellCommander = func(name string, arg ...string) command.IShellCommand {
			return internal.TestShellCommand{
				OutputterFunc: func() ([]byte, error) {
					return []byte("foobar"), nil
				},
			}
		}

		out, err := drupal.Drush("", "local", []string{"status"}).Exec()
		assert.NoError(err)
		assert.Equal([]byte("foobar"), out)
	})

}

func TestDrushQuery(t *testing.T) {
	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	command.ShellCommander = func(name string, arg ...string) command.IShellCommand {
		return internal.TestShellCommand{
			OutputterFunc: func() ([]byte, error) {
				return []byte(strings.Join(arg, ",")), nil
			},
		}
	}

	cmd, err := drupal.Drush("", "", []string{}).Query("SELECT uid FROM users")
	assert.NoError(t, err)
	assert.Equal(t, []byte("sql:query,SELECT uid FROM users"), cmd)
}
