package drupal

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"salsadigitalauorg/shipshape/pkg/shipshape"
	"salsadigitalauorg/shipshape/pkg/utils"
	"sort"
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

// UnmarshalDataMap parses the drush permissions yaml into the DrushRoles
// type for further processing.
func (c *DbPermissionsCheck) UnmarshalDataMap() {
	if len(c.DataMap[c.ConfigName]) == 0 {
		c.AddFail("no data provided")
	}

	c.Permissions = map[string]DrushRole{}
	err := yaml.Unmarshal([]byte(c.DataMap[c.ConfigName]), &c.Permissions)
	if err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			c.AddFail(err.Error())
			return
		}
	}
}

// RunCheck implements the Check logic for Drupal Permissions in database config.
func (c *DbPermissionsCheck) RunCheck() {
	if len(c.Disallowed) == 0 {
		c.AddFail("list of disallowed perms not provided")
	}

	for r, perms := range c.Permissions {
		fails := utils.StringSlicesIntersect(perms.Perms, c.Disallowed)
		if len(fails) == 0 {
			c.AddPass(fmt.Sprintf("[%s] no disallowed permissions", r))
			continue
		}

		// Sort fails.
		sort.Slice(fails, func(i int, j int) bool {
			return fails[i] < fails[j]
		})
		c.AddFail(fmt.Sprintf("[%s] disallowed permissions: [%s]", r, strings.Join(fails, ", ")))
	}

	if len(c.Result.Failures) == 0 {
		c.Result.Status = shipshape.Pass
	}
}
