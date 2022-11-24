package drupal

import (
	"os/exec"
	"path/filepath"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const DrushDefaultPath = "vendor/drush/drush/drush"

var ExecCommand = exec.Command

// Drush is a simple wrapper around DrushCommand which allows chaining
// commands for Drush, e.g, `Drush("", "", "status").Exec()`.
func Drush(drushPath string, alias string, command []string) *DrushCommand {
	if drushPath == "" {
		drushPath = DrushDefaultPath
	}
	if !filepath.IsAbs(drushPath) {
		drushPath = filepath.Join(shipshape.ProjectDir, drushPath)
	}
	return &DrushCommand{DrushPath: drushPath, Alias: alias, Args: command}
}

// Merge implementation for DrushCommand.
func (cmd *DrushCommand) Merge(mergeCmd DrushCommand) {
	utils.MergeString(&cmd.DrushPath, mergeCmd.DrushPath)
	utils.MergeString(&cmd.Alias, mergeCmd.Alias)
	utils.MergeStringSlice(&cmd.Args, mergeCmd.Args)
}

// Exec runs the drush command and returns the output.
func (cmd *DrushCommand) Exec() ([]byte, error) {
	if cmd.Alias != "" {
		cmd.Args = append([]string{"@" + cmd.Alias}, cmd.Args...)
	}
	return ExecCommand(cmd.DrushPath, cmd.Args...).Output()
}

// Query runs the drush sql:query command and returns the output.
func (cmd *DrushCommand) Query(qry string) ([]byte, error) {
	cmd.Args = []string{"sql:query", qry}
	return cmd.Exec()
}
