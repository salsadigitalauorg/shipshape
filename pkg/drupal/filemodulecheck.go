package drupal

import (
	"path/filepath"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// Merge implementation for file check.
func (c *FileModuleCheck) Merge(mergeCheck shipshape.Check) error {
	fileModuleMergeCheck := mergeCheck.(*FileModuleCheck)
	if err := c.YamlCheck.Merge(&fileModuleMergeCheck.YamlCheck); err != nil {
		return err
	}

	utils.MergeStringSlice(&c.Required, fileModuleMergeCheck.Required)
	utils.MergeStringSlice(&c.Disallowed, fileModuleMergeCheck.Disallowed)
	return nil
}

// RunCheck applies the Check logic for Drupal Modules in config files.
func (c *FileModuleCheck) RunCheck(remediate bool) {
	configName := filepath.Join(c.Path, c.File)
	CheckModulesInYaml(&c.YamlBase, FileModule, configName, c.Required, c.Disallowed)
}

// Init implementation for the File-based module check.
func (c *FileModuleCheck) Init(ct shipshape.CheckType) {
	c.CheckBase.Init(ct)
	c.File = "core.extension.yml"
	if c.IgnoreMissing == nil {
		cTrue := true
		c.IgnoreMissing = &cTrue
	}
}
