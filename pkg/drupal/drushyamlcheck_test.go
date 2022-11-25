package drupal_test

import (
	"os/exec"
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestDrushYamlMerge(t *testing.T) {
	assert := assert.New(t)

	c := drupal.DrushYamlCheck{
		YamlBase: shipshape.YamlBase{
			Values: []shipshape.KeyValue{
				{Key: "key1", Value: "val1", Optional: false},
			},
		},
		DrushCommand: drupal.DrushCommand{
			DrushPath: "/path/to/drush",
			Alias:     "alias1",
			Args:      []string{"arg1"},
		},
		Command:    "command1",
		ConfigName: "configname1",
	}
	c.Merge(&drupal.DrushYamlCheck{
		YamlBase: shipshape.YamlBase{
			Values: []shipshape.KeyValue{
				{Key: "key1", Value: "val1", Optional: true},
			},
		},
		DrushCommand: drupal.DrushCommand{
			DrushPath: "/new/path/to/drush",
			Alias:     "alias2",
			Args:      []string{"arg2"},
		},
		Command: "command2",
	})
	assert.EqualValues(drupal.DrushYamlCheck{
		YamlBase: shipshape.YamlBase{
			Values: []shipshape.KeyValue{
				{Key: "key1", Value: "val1", Optional: true},
			},
		},
		DrushCommand: drupal.DrushCommand{
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

	c := drupal.DrushYamlCheck{
		Command:    "status",
		ConfigName: "core.extension",
	}

	c.Init(drupal.DrushYaml)
	assert.True(c.RequiresDb)

	c.FetchData()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.ElementsMatch(
		[]string{"vendor/drush/drush/drush: no such file or directory"},
		c.Result.Failures,
	)

	c = drupal.DrushYamlCheck{
		Command:    "status",
		ConfigName: "core.extension",
	}
	drupal.ExecCommand = internal.FakeExecCommand
	defer func() { drupal.ExecCommand = exec.Command }()
	internal.MockedExitStatus = 1
	internal.MockedStderr = "unable to run drush command"
	c.FetchData()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.ElementsMatch(
		[]string{"unable to run drush command"},
		c.Result.Failures,
	)

	internal.MockedExitStatus = 0
	internal.MockedStdout = `
module:
  block: 0
  views_ui: 0

`
	c = drupal.DrushYamlCheck{
		Command:    "status",
		ConfigName: "core.extension",
	}
	c.FetchData()
	assert.NotEqual(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.Empty(c.Result.Failures)
}
