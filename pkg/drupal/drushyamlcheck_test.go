package drupal_test

import (
	"os/exec"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	. "github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"

	"github.com/stretchr/testify/assert"
)

func TestDrushYamlMerge(t *testing.T) {
	assert := assert.New(t)

	c := DrushYamlCheck{
		YamlBase: shipshape.YamlBase{
			Values: []shipshape.KeyValue{
				{Key: "key1", Value: "val1", Optional: false},
			},
		},
		DrushCommand: DrushCommand{
			DrushPath: "/path/to/drush",
			Alias:     "alias1",
			Args:      []string{"arg1"},
		},
		Command:    "command1",
		ConfigName: "configname1",
	}
	c.Merge(&DrushYamlCheck{
		YamlBase: shipshape.YamlBase{
			Values: []shipshape.KeyValue{
				{Key: "key1", Value: "val1", Optional: true},
			},
		},
		DrushCommand: DrushCommand{
			DrushPath: "/new/path/to/drush",
			Alias:     "alias2",
			Args:      []string{"arg2"},
		},
		Command: "command2",
	})
	assert.EqualValues(DrushYamlCheck{
		YamlBase: shipshape.YamlBase{
			Values: []shipshape.KeyValue{
				{Key: "key1", Value: "val1", Optional: true},
			},
		},
		DrushCommand: DrushCommand{
			DrushPath: "/new/path/to/drush",
			Alias:     "alias2",
			Args:      []string{"arg2"},
		},
		Command:    "command2",
		ConfigName: "configname1",
	}, c)
}

func TestDrushYamlCheck(t *testing.T) {
	assert := assert.New(t)

	t.Run("drushNotFound", func(t *testing.T) {
		c := DrushYamlCheck{
			Command:    "status",
			ConfigName: "core.extension",
		}

		c.Init(DrushYaml)
		assert.True(c.RequiresDb)

		c.FetchData()
		assert.Equal(config.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.ElementsMatch(
			[]string{"vendor/drush/drush/drush: no such file or directory"},
			c.Result.Failures,
		)
	})

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	t.Run("drushError", func(t *testing.T) {
		c := DrushYamlCheck{
			Command:    "status",
			ConfigName: "core.extension",
		}

		command.ShellCommander = internal.ShellCommanderMaker(
			nil,
			&exec.ExitError{Stderr: []byte("unable to run drush command")},
			nil,
		)

		c.FetchData()
		assert.Equal(config.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.ElementsMatch(
			[]string{"unable to run drush command"},
			c.Result.Failures,
		)
	})

	t.Run("drushOK", func(t *testing.T) {
		stdout := `
module:
  block: 0
  views_ui: 0

`

		command.ShellCommander = internal.ShellCommanderMaker(
			&stdout,
			nil,
			nil,
		)

		c := DrushYamlCheck{
			Command:    "status",
			ConfigName: "core.extension",
		}
		c.FetchData()
		assert.NotEqual(config.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.Empty(c.Result.Failures)
	})
}
