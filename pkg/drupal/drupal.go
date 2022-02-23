package drupal

import (
	"fmt"
	"salsadigitalauorg/shipshape/pkg/core"
)

// RunCheck applies the Check logic for Drupal Modules in yaml content.
// It uses YamlBase to verify that the list of provided Required or
// Disallowed modules are installed or not.
func CheckModulesInYaml(c *core.YamlBase, ct core.CheckType, configName string, required []string, disallowed []string) {
	moduleKey := func(m string) core.KeyValue {
		if ct == FileModule {
			return core.KeyValue{
				Key:   "module." + m,
				Value: "0",
			}
		}
		return core.KeyValue{
			Key:   m + ".status",
			Value: "Enabled",
		}
	}

	for _, m := range required {
		kvr, _, err := c.CheckKeyValue(moduleKey(m), configName)
		// It could be a value different from 0, which still means it's enabled.
		if kvr == core.KeyValueEqual || kvr == core.KeyValueNotEqual {
			c.AddPass(fmt.Sprintf("'%s' is enabled", m))
		} else if kvr == core.KeyValueError {
			c.AddFail(err.Error())
		} else {
			c.AddFail(fmt.Sprintf("'%s' is not enabled", m))
		}
	}
	for _, m := range disallowed {
		kvr, _, err := c.CheckKeyValue(moduleKey(m), configName)
		// It could be a value different from 0, which still means it's enabled.
		if kvr == core.KeyValueEqual || kvr == core.KeyValueNotEqual {
			c.AddFail(fmt.Sprintf("'%s' is enabled", m))
		} else if kvr == core.KeyValueError {
			c.AddFail(err.Error())
		} else {
			c.AddPass(fmt.Sprintf("'%s' is not enabled", m))
		}
	}

	if len(c.Result.Failures) > 0 {
		c.Result.Status = core.Fail
	} else {
		c.Result.Status = core.Pass
	}
}

// RunCheck applies the Check logic for Drupal Modules in config files.
func (c *FileModuleCheck) RunCheck() {
	CheckModulesInYaml(&c.YamlBase, FileModule, c.File, c.Required, c.Disallowed)
}

// Init implementation for the File-based module check.
func (c *FileModuleCheck) Init(pd string, ct core.CheckType) {
	c.CheckBase.Init(pd, ct)
	c.File = "core.extension.yml"
	c.IgnoreMissing = true
}
