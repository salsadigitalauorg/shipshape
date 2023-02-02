package drupal_test

import (
	"errors"
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
		command.ShellCommander = internal.ShellCommanderMaker(
			nil, errors.New("bash: drushfoo: command not found"), nil)
		_, err := drupal.Drush("", "", []string{"status"}).Exec()
		assert.Error(err, "bash: drushfoo: command not found")
	})

	t.Run("ok", func(t *testing.T) {
		command.ShellCommander = internal.ShellCommanderMaker(&[]string{"foobar"}[0], nil, nil)
		out, err := drupal.Drush("", "local", []string{"status"}).Exec()
		assert.NoError(err)
		assert.Equal([]byte("foobar"), out)
	})

}

func TestDrushQuery(t *testing.T) {
	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	var generatedCommand string
	command.ShellCommander = internal.ShellCommanderMaker(nil, nil, &generatedCommand)

	_, err := drupal.Drush("", "", []string{}).Query("SELECT uid FROM users")
	assert.NoError(t, err)
	assert.Equal(t, "vendor/drush/drush/drush sql:query SELECT uid FROM users", generatedCommand)
}
