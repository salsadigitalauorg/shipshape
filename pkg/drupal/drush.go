package drupal

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"salsadigitalauorg/shipshape/pkg/shipshape"
	"strings"

	"gopkg.in/yaml.v2"
)

const DrushDefaultPath = "vendor/drush/drush/drush"

var ExecCommand = exec.Command

// Drush is a simple wrapper around DrushCommand which allows chaining
// commands for Drush, e.g, `Drush("", "", "status").Exec()`.
func Drush(drushPath string, alias string, command string) *DrushCommand {
	if drushPath == "" {
		drushPath = DrushDefaultPath
	}
	if !filepath.IsAbs(drushPath) {
		drushPath = filepath.Join(shipshape.ProjectDir, drushPath)
	}
	return &DrushCommand{DrushPath: drushPath, Alias: alias, Command: command}
}

// Exec runs the drush command and returns the output.
func (cmd *DrushCommand) Exec() ([]byte, error) {
	cmdSlice := strings.Split(cmd.Command, " ")
	if cmd.Alias != "" {
		cmdSlice = append([]string{"@" + cmd.Alias}, cmdSlice...)
	}
	return ExecCommand(cmd.DrushPath, cmdSlice...).Output()
}

// FetchData runs the drush command to populate data for the Drush Yaml check.
// Since the check is going to be Yaml-based, `--format=yaml` is automatically
// added to the command.
func (c *DrushYamlCheck) FetchData() {
	var err error
	c.DataMap = map[string][]byte{}
	c.DataMap[c.ConfigName], err = Drush(c.DrushPath, c.Alias, c.Command+" --format=yaml").Exec()
	if err != nil {
		if pathErr, ok := err.(*fs.PathError); ok {
			c.AddFail(pathErr.Path + ": " + pathErr.Err.Error())
		} else {
			c.AddFail(string(err.(*exec.ExitError).Stderr))
		}
	}
}

// Init implementation for the DB-based module check.
func (c *DbModuleCheck) Init(pd string, ct shipshape.CheckType) {
	c.CheckBase.Init(pd, ct)
	c.Command = "pm:list --status=enabled"
}

// RunCheck applies the Check logic for Drupal Modules in database config.
func (c *DbModuleCheck) RunCheck() {
	CheckModulesInYaml(&c.YamlBase, DbModule, c.ConfigName, c.Required, c.Disallowed)
}

// Init implementation for the DB-based permissions check.
func (c *DbPermissionsCheck) Init(pd string, ct shipshape.CheckType) {
	c.CheckBase.Init(pd, ct)
	c.Command = "role:list"
	c.ConfigName = "permissions"
}

// UnmarshalDataMap parses the drush permissions yaml into the DrushPermissions
// type for further processing.
func (c *DbPermissionsCheck) UnmarshalDataMap() {
	c.Permissions = DrushPermissions{}
	err := yaml.Unmarshal([]byte(c.DataMap[c.ConfigName]), &c.Permissions)
	if err != nil {
		c.AddFail(err.Error())
		return
	}
}

// RunCheck implements the Check logic for Drupal Permissions in database config.
func (c *DbPermissionsCheck) RunCheck() {

}
