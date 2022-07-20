package drupal_test

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
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

	c.Init("", drupal.DrushYaml)
	if !c.RequiresDb {
		t.Error("expected RequiresDb to be true, got false")
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
				YamlBase: shipshape.YamlBase{
					CheckBase: shipshape.CheckBase{DataMap: dataMap},
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

func TestDbPermissionsCheck(t *testing.T) {
	// Test init.
	c := drupal.DbPermissionsCheck{}
	c.Init("", drupal.DbPermissions)
	if c.Command != "role:list" {
		t.Errorf("Command should be 'role:list', got %s", c.Command)
	}
	if c.ConfigName != "permissions" {
		t.Errorf("ConfigName should be 'permissions', got %s", c.ConfigName)
	}
	c.UnmarshalDataMap()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"no data provided"}); !ok {
		t.Error(msg)
	}

	// Test UnmarshalDataMap.
	c = drupal.DbPermissionsCheck{}
	c.Init("", drupal.DbPermissions)
	c.DataMap = map[string][]byte{
		"permissions": []byte(`
site_administrator:
  label: 'Site Administrator'
  perms: {  }
anonymous:
  label: 'Anonymous user'
  perms:
    - 'access content'
    - 'view media'
authenticated:
  label: 'Authenticated user'
  perms:
    - 'access content'
    - 'view media'
site_editor:
  label: 'Site Editor'
  perms: {  }
`),
	}
	c.UnmarshalDataMap()
	if msg, ok := internal.EnsureNoFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if c.Permissions == nil {
		t.Errorf("Permissions should be populated")
	}
	expectedPermissions := map[string]drupal.DrushRole{
		"anonymous": {
			Label: "Anonymous user",
			Perms: []string{"access content", "view media"},
		},
		"authenticated": {
			Label: "Authenticated user",
			Perms: []string{"access content", "view media"},
		},
		"site_administrator": {
			Label: "Site Administrator",
			Perms: []string(nil),
		},
		"site_editor": {
			Label: "Site Editor",
			Perms: []string(nil),
		},
	}
	if !reflect.DeepEqual(c.Permissions, expectedPermissions) {
		t.Errorf("Permissions are not as expected, got: %#v", c.Permissions)
	}

	// Test RunCheck.
	c = drupal.DbPermissionsCheck{}
	c.Init("", drupal.DbPermissions)
	c.RunCheck()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"list of disallowed perms not provided"}); !ok {
		t.Error(msg)
	}

	c = drupal.DbPermissionsCheck{}
	c.Init("", drupal.DbPermissions)
	c.Permissions = map[string]drupal.DrushRole{
		"anonymous": {
			Label: "Anonymous user",
			Perms: []string{"access content", "view media"},
		},
		"authenticated": {
			Label: "Authenticated user",
			Perms: []string{"access content", "view media"},
		},
		"site_administrator": {
			Label: "Site Administrator",
			Perms: []string(nil),
		},
		"site_editor": {
			Label: "Site Editor",
			Perms: []string(nil),
		},
	}
	c.Disallowed = []string{"administer modules"}
	c.RunCheck()
	c.Result.Sort()
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"[anonymous] no disallowed permissions",
		"[authenticated] no disallowed permissions",
		"[site_administrator] no disallowed permissions",
		"[site_editor] no disallowed permissions",
	}); !ok {
		t.Error(msg)
	}

	c = drupal.DbPermissionsCheck{}
	c.Init("", drupal.DbPermissions)
	c.Permissions = map[string]drupal.DrushRole{
		"anonymous": {
			Label: "Anonymous user",
			Perms: []string{"access content", "view media"},
		},
		"authenticated": {
			Label: "Authenticated user",
			Perms: []string{"access content", "view media"},
		},
		"site_administrator": {
			Label: "Site Administrator",
			Perms: []string{"administer modules", "administer permissions"},
		},
		"another_site_administrator": {
			Label: "Site Administrator",
			Perms: []string{"administer modules", "administer permissions"},
		},
		"site_editor": {
			Label: "Site Editor",
			Perms: []string{"administer modules"},
		},
	}
	c.Disallowed = []string{"administer modules", "administer permissions"}
	c.ExcludeRoles = []string{"another_site_administrator"}
	c.RunCheck()
	c.Result.Sort()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"[anonymous] no disallowed permissions",
		"[authenticated] no disallowed permissions",
	}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{
		"[site_administrator] disallowed permissions: [administer modules, administer permissions]",
		"[site_editor] disallowed permissions: [administer modules]",
	}); !ok {
		t.Error(msg)
	}
}

func TestTrackingCodeCheckFails(t *testing.T) {
	c := drupal.TrackingCodeCheck{
		Code: "UA-xxxxxx-1",
	}
	c.Init("", drupal.TrackingCode)
	if c.Command != "status" {
		t.Error("drush command for check should already be set")
	}

	c.DrushStatus = drupal.DrushStatus{
		Uri: "https://google.com",
	}
	c.RunCheck()

	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{
		"tracking code [UA-xxxxxx-1] not present",
	}); !ok {
		t.Error(msg)
	}
}

func TestTrackingCodeCheckPass(t *testing.T) {
	c := drupal.TrackingCodeCheck{
		Code: "UA-xxxxxx-1",
	}
	c.Init("", drupal.TrackingCode)
	if c.Command != "status" {
		t.Error("drush command for check should already be set")
	}

	c.DrushStatus = drupal.DrushStatus{
		Uri: "https://gist.github.com/Pominova/cf7884e7418f6ebfa412d2d3dc472a97",
	}
	c.RunCheck()

	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"tracking code [UA-xxxxxx-1] present",
	}); !ok {
		t.Error(msg)
	}
}
