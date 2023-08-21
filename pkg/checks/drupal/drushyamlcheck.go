package drupal

import (
	"io/fs"
	"os/exec"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// Init implementation for the drush-based yaml check.
func (c *DrushYamlCheck) Init(ct config.CheckType) {
	c.YamlBase.Init(ct)
	c.RequiresDb = true
}

// Merge implementation for DbModuleCheck check.
func (c *DrushYamlCheck) Merge(mergeCheck config.Check) error {
	drushYamlMergeCheck := mergeCheck.(*DrushYamlCheck)
	if err := c.YamlBase.Merge(&drushYamlMergeCheck.YamlBase); err != nil {
		return err
	}

	c.DrushCommand.Merge(drushYamlMergeCheck.DrushCommand)
	utils.MergeString(&c.Command, drushYamlMergeCheck.Command)
	utils.MergeString(&c.ConfigName, drushYamlMergeCheck.ConfigName)
	return nil
}

// FetchData runs the drush command to populate data for the Drush Yaml check.
// Since the check is going to be Yaml-based, `--format=yaml` is automatically
// added to the command.
func (c *DrushYamlCheck) FetchData() {
	var err error
	c.DataMap = map[string][]byte{}
	c.DrushCommand.Args = append(strings.Fields(c.Command), "--format=yaml")
	c.DataMap[c.ConfigName], err = Drush(c.DrushPath, c.Alias, c.DrushCommand.Args).Exec()
	if err != nil {
		if pathErr, ok := err.(*fs.PathError); ok {
			c.AddFail(pathErr.Path + ": " + pathErr.Err.Error())
			c.AddBreach(result.ValueBreach{
				Value: pathErr.Path + ": " + pathErr.Err.Error()})
		} else {
			msg := string(err.(*exec.ExitError).Stderr)
			c.AddFail(strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", ""))
			c.AddBreach(result.ValueBreach{
				Value: strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", "")})
		}
	}
}
