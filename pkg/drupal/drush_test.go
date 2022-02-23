package drupal_test

import (
	"fmt"
	"os"
	"os/exec"
	"salsadigitalauorg/shipshape/internal"
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"
	"strconv"
	"testing"
)

var mockedExitStatus = 0
var mockedStdout string
var mockedStderr string

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	es := strconv.Itoa(mockedExitStatus)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		"STDERR=" + mockedStderr,
		"EXIT_STATUS=" + es}
	return cmd
}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	if i > 0 {
		fmt.Fprint(os.Stderr, os.Getenv("STDERR"))
	} else {
		fmt.Fprint(os.Stdout, os.Getenv("STDOUT"))
	}
	os.Exit(i)
}

func TestDrushExec(t *testing.T) {
	drupal.ExecCommand = fakeExecCommand
	defer func() { drupal.ExecCommand = exec.Command }()

	// Command not found.
	mockedExitStatus = 127
	mockedStderr = "bash: drushfoo: command not found"
	_, err := drupal.Drush("", "", "status").Exec()
	if err == nil || string(err.(*exec.ExitError).Stderr) != "bash: drushfoo: command not found" {
		t.Errorf("Drush call should fail, got %#v", err)
	}

	mockedExitStatus = 0
	mockedStdout = "foobar"
	_, err = drupal.Drush("", "local", "status").Exec()
	if err != nil {
		t.Errorf("Drush call should pass, got %#v", err)
	}
}

func TestDrushYamlCheck(t *testing.T) {
	c := drupal.DrushYamlCheck{
		DrushCommand: drupal.DrushCommand{Command: "status"},
		ConfigName:   "core.extension",
	}
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"vendor/drush/drush/drush: no such file or directory"}); !ok {
		t.Error(msg)
	}

	c = drupal.DrushYamlCheck{
		DrushCommand: drupal.DrushCommand{Command: "status"},
		ConfigName:   "core.extension",
	}
	drupal.ExecCommand = fakeExecCommand
	defer func() { drupal.ExecCommand = exec.Command }()
	mockedExitStatus = 1
	mockedStderr = "unable to run drush command"
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"unable to run drush command"}); !ok {
		t.Error(msg)
	}

	mockedExitStatus = 0
	mockedStdout = `
module:
  block: 0
  views_ui: 0

`
	c = drupal.DrushYamlCheck{
		DrushCommand: drupal.DrushCommand{Command: "status"},
		ConfigName:   "core.extension",
	}
	c.FetchData()
	if msg, ok := internal.EnsureNoFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
}

func TestDbModuleCheck(t *testing.T) {
	c := drupal.DbModuleCheck{}
	c.Init("", drupal.DbModule)
	if c.Command != "pm:list --status=enabled" {
		t.Errorf("drush command for check should be already set")
	}

	mockCheck := func(dataMap map[string][]byte) drupal.DbModuleCheck {
		if dataMap == nil {
			dataMap = map[string][]byte{
				"modules": []byte(`
block:
  status: enabled
node:
  status: enabled

`),
			}
		}
		c := drupal.DbModuleCheck{
			DrushYamlCheck: drupal.DrushYamlCheck{
				YamlBase: core.YamlBase{
					CheckBase: core.CheckBase{DataMap: dataMap},
				},
				ConfigName: "modules",
			},
			Required:   []string{"block", "node"},
			Disallowed: []string{"views_ui", "field_ui"},
		}
		c.Init("", drupal.DbModule)
		c.UnmarshalDataMap()
		c.RunCheck()
		return c
	}

	c = mockCheck(nil)
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"'block' is enabled",
		"'node' is enabled",
		"'views_ui' is not enabled",
		"'field_ui' is not enabled",
	}); !ok {
		t.Error(msg)
	}

	c = mockCheck(map[string][]byte{
		"modules": []byte(`
node:
  status: enabled
views_ui:
  status: enabled

`),
	})

	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"'node' is enabled",
		"'field_ui' is not enabled",
	}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{
		"'block' is not enabled",
		"'views_ui' is enabled",
	}); !ok {
		t.Error(msg)
	}
}
