package json

import (
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

func RegisterChecks() {
	config.ChecksRegistry[Json] = func() config.Check { return &JsonCheck{} }
}

func init() {
	RegisterChecks()
}

// Merge implementation for JSON check.
func (c *JsonCheck) Merge(mergeCheck config.Check) error {
	yCheck := mergeCheck.(*JsonCheck)
	if err := c.YamlCheck.Merge(&yCheck.YamlCheck); err != nil {
		return err
	}

	return nil
}

// RunCheck implements the base logic for running checks against Yaml data.
func (c *JsonCheck) RunCheck() {
	for configName := range c.DataMap {
		c.processData(configName)
	}
}
