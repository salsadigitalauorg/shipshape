package drupal_test

import (
	"os/exec"
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/stretchr/testify/assert"
)

func TestExecCommandHelper(t *testing.T) {
	internal.TestExecCommandHelper(t)
}

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
	drupal.ExecCommand = internal.FakeExecCommand
	defer func() { drupal.ExecCommand = exec.Command }()

	// Command not found.
	internal.MockedExitStatus = 127
	internal.MockedStderr = "bash: drushfoo: command not found"
	_, err := drupal.Drush("", "", []string{"status"}).Exec()
	if err == nil || string(err.(*exec.ExitError).Stderr) != "bash: drushfoo: command not found" {
		t.Errorf("Drush call should fail, got %#v", err)
	}

	internal.MockedExitStatus = 0
	internal.MockedStdout = "foobar"
	_, err = drupal.Drush("", "local", []string{"status"}).Exec()
	if err != nil {
		t.Errorf("Drush call should pass, got %#v", err)
	}
}

func TestDrushQuery(t *testing.T) {
	drupal.ExecCommand = internal.FakeExecCommand
	defer func() { drupal.ExecCommand = exec.Command }()
	drupal.Drush("", "", []string{}).Query("SELECT uid FROM users")
	assert.EqualValues(t, []string{"sql:query", "SELECT uid FROM users"}, internal.FakeCommandArgs)
}
