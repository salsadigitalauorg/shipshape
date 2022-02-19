package drupal

import (
	"fmt"
	"salsadigitalauorg/shipshape/pkg/core"
)

func (c *FileModuleCheck) RunCheck() {
	if c.DataMap == nil {
		c.Result.Status = core.Fail
		c.Result.Failures = append(
			c.Result.Failures,
			"no data to run check on",
		)
		return
	}

	err := c.UnmarshalDataMap()
	if err != nil {
		c.Result.Status = core.Fail
		c.Result.Failures = append(
			c.Result.Failures,
			err.Error(),
		)
		return
	}

	for _, m := range c.Required {
		kvr, _, err := c.CheckKeyValue(core.KeyValue{
			Key:   "module." + m,
			Value: "0",
		}, c.File+".yml")
		// It could be a value different from 0, which still means it's enabled.
		if kvr == core.KeyValueEqual || kvr == core.KeyValueNotEqual {
			c.Result.Passes = append(
				c.Result.Passes,
				fmt.Sprintf("'%s' is enabled", m),
			)
			if c.Result.Status == "" {
				c.Result.Status = core.Pass
			}
		} else if kvr == core.KeyValueError {
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
			c.Result.Status = core.Fail
		} else {
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("'%s' is not enabled", m),
			)
			c.Result.Status = core.Fail
		}
	}
	for _, m := range c.Disallowed {
		kvr, _, err := c.CheckKeyValue(core.KeyValue{
			Key:   "module." + m,
			Value: "0",
		}, c.File+".yml")
		// It could be a value different from 0, which still means it's enabled.
		if kvr == core.KeyValueEqual || kvr == core.KeyValueNotEqual {
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("'%s' is enabled", m),
			)
			c.Result.Status = core.Fail
		} else if kvr == core.KeyValueError {
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
			c.Result.Status = core.Fail
		} else {
			c.Result.Passes = append(
				c.Result.Passes,
				fmt.Sprintf("'%s' is not enabled", m),
			)
			if c.Result.Status == "" {
				c.Result.Status = core.Pass
			}
		}
	}
	c.YamlCheck.RunCheck()
}

func (c *FileModuleCheck) Init(pd string, ct core.CheckType) {
	c.CheckBase.Init(pd, ct)
	c.File = "core.extension"
}

func (c *DbModuleCheck) RunCheck() {
	if c.DataMap == nil {
		c.Result.Status = core.Fail
		c.Result.Failures = append(
			c.Result.Failures,
			"no data to run check on",
		)
		return
	}
}
