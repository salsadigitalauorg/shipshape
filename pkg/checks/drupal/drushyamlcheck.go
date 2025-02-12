package drupal

import (
	"fmt"
	"io/fs"
	"os/exec"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/remediation"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	log "github.com/sirupsen/logrus"
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
			c.AddBreach(&breach.ValueBreach{
				Value: pathErr.Path + ": " + pathErr.Err.Error()})
		} else {
			msg := string(err.(*exec.ExitError).Stderr)
			c.AddBreach(&breach.ValueBreach{
				ValueLabel: c.ConfigName,
				Value:      strings.ReplaceAll(strings.TrimSpace(msg), "  \n  ", "")})
		}
	}
}

// Remediate attempts to remediate a breach by running the drush command
// specified in the check.
func (c *DrushYamlCheck) Remediate() {
	for _, b := range c.Result.Breaches {
		contextLogger := log.WithFields(log.Fields{
			"check-type": c.GetType(),
			"check-name": c.GetName(),
			"breach":     b,
		})
		if c.RemediateCommand == "" {
			contextLogger.Print("no remediation command specified - failing")
			b.SetRemediation(remediation.RemediationStatusNoSupport, "")
			return
		}

		contextLogger.Print("running remediation command")
		_, err := command.ShellCommander("sh", "-c", c.RemediateCommand).Output()
		if err != nil {
			b.SetRemediation(remediation.RemediationStatusFailed, fmt.Sprintf(
				"error running remediation command for config '%s' due to error: %s",
				c.ConfigName, command.GetMsgFromCommandError(err)))
		} else {
			if c.RemediateMsg == "" {
				c.RemediateMsg = fmt.Sprintf(
					"remediation command for config '%s' ran successfully", c.ConfigName)
			}
			b.SetRemediation(remediation.RemediationStatusSuccess, c.RemediateMsg)
		}
	}
}
