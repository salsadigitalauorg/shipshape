package drupal

import (
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// Init implementation for the DB-based module check.
func (c *DbModuleCheck) Init(ct shipshape.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
	c.Command = "pm:list --status=enabled"
}

// Merge implementation for DbModuleCheck check.
func (c *DbModuleCheck) Merge(mergeCheck shipshape.Check) error {
	dbModuleMergeCheck := mergeCheck.(*DbModuleCheck)
	if err := c.DrushYamlCheck.Merge(&dbModuleMergeCheck.DrushYamlCheck); err != nil {
		return err
	}

	utils.MergeStringSlice(&c.Required, dbModuleMergeCheck.Required)
	utils.MergeStringSlice(&c.Disallowed, dbModuleMergeCheck.Disallowed)
	return nil
}

// RunCheck applies the Check logic for Drupal Modules in database config.
func (c *DbModuleCheck) RunCheck(remediate bool) {
	CheckModulesInYaml(&c.YamlBase, DbModule, c.ConfigName, c.Required, c.Disallowed)
}
