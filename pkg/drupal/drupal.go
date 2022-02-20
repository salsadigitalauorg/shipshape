package drupal

import (
	"fmt"
	"salsadigitalauorg/shipshape/pkg/core"
)

// RunCheck applies the Check logic for Drupal Modules in config files.
// It uses YamlCheck as a base to verify that the list of provided Required or
// Disallowed modules are installed or not.
func (c *FileModuleCheck) RunCheck() {
	for _, m := range c.Required {
		kvr, _, err := c.CheckKeyValue(core.KeyValue{
			Key:   "module." + m,
			Value: "0",
		}, c.File+".yml")
		// It could be a value different from 0, which still means it's enabled.
		if kvr == core.KeyValueEqual || kvr == core.KeyValueNotEqual {
			c.AddPass(fmt.Sprintf("'%s' is enabled", m))
		} else if kvr == core.KeyValueError {
			c.AddFail(err.Error())
		} else {
			c.AddFail(fmt.Sprintf("'%s' is not enabled", m))
		}
	}
	for _, m := range c.Disallowed {
		kvr, _, err := c.CheckKeyValue(core.KeyValue{
			Key:   "module." + m,
			Value: "0",
		}, c.File+".yml")
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

func (c *FileModuleCheck) Init(pd string, ct core.CheckType) {
	c.CheckBase.Init(pd, ct)
	c.File = "core.extension"
}

func (c *DbModuleCheck) RunCheck() {}
