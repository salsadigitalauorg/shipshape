package yaml

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

// Merge implementation for Yaml check.
func (c *YamlLintCheck) Merge(mergeCheck config.Check) error {
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
		var ifc interface{}
		err := yaml.Unmarshal([]byte(data), &ifc)
		if err != nil {
			if typeErr, ok := err.(*yaml.TypeError); ok {
				c.AddBreach(&breach.ValueBreach{
					ValueLabel: "cannot decode yaml: " + f,
					Value:      strings.Join(typeErr.Errors, "\n")})
			} else {
				c.AddBreach(&breach.ValueBreach{
					ValueLabel: "yaml error: " + f,
					Value:      err.Error()})
			}
		} else {
			c.AddPass(fmt.Sprintf("%s has valid yaml.", f))
		}
	}
	if c.Result.Status != result.Fail {
		c.Result.Status = result.Pass
	}
}

// RunCheck for YamlLint does nothing since the check is in UnmarshalDataMap.
func (c *YamlLintCheck) RunCheck() {}
