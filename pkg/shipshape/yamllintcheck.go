package shipshape

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Merge implementation for Yaml check.
func (c *YamlLintCheck) Merge(mergeCheck Check) error {
	yamlLintMergeCheck := mergeCheck.(*YamlLintCheck)
	if err := c.YamlCheck.Merge(&yamlLintMergeCheck.YamlCheck); err != nil {
		return err
	}
	return nil
}

// UnmarshalDataMap tries to parse the yaml file into a generic structure and
// returns any errors as failures.
func (c *YamlLintCheck) UnmarshalDataMap() {
	for f, data := range c.DataMap {
		ifc := make(map[string]interface{})
		err := yaml.Unmarshal([]byte(data), &ifc)
		if err != nil {
			if typeErr, ok := err.(*yaml.TypeError); ok {
				for _, msg := range typeErr.Errors {
					c.AddFail(fmt.Sprintf("[%s] %s", f, msg))
				}
			} else {
				c.AddFail(fmt.Sprintf("[%s] %s", f, err.Error()))
			}
		} else {
			c.AddPass(fmt.Sprintf("%s has valid yaml.", f))
		}
	}
	if c.Result.Status != Fail {
		c.Result.Status = Pass
	}
}

// RunCheck for YamlLint does nothing since the check is in UnmarshalDataMap.
func (c *YamlLintCheck) RunCheck() {}
