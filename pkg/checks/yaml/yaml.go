package yaml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	"gopkg.in/yaml.v3"
)

//go:generate go run ../../../cmd/gen.go registry --checkpackage=yaml

func RegisterChecks() {
	config.ChecksRegistry[Yaml] = func() config.Check { return &YamlCheck{} }
	config.ChecksRegistry[YamlLint] = func() config.Check { return &YamlLintCheck{} }
}

func init() {
	RegisterChecks()
}

// Merge implementation for YamlBase.
func (c *YamlBase) Merge(mergeCheck config.Check) error {
	yBaseCheck := mergeCheck.(*YamlBase)
	if err := c.CheckBase.Merge(&yBaseCheck.CheckBase); err != nil {
		return err
	}

	MergeKeyValueSlice(&c.Values, yBaseCheck.Values)
	return nil
}

// RunCheck implements the base logic for running checks against Yaml data.
func (c *YamlBase) RunCheck() {
	for configName := range c.DataMap {
		c.processData(configName)
	}
}

// UnmarshalDataMap parses the DataMap into Yaml for further processing.
// DataMap is expected to be populated from FetchData in the respective Check
// implementation.
func (c *YamlBase) UnmarshalDataMap() {
	c.NodeMap = map[string]yaml.Node{}
	for configName, data := range c.DataMap {
		n := yaml.Node{}
		err := yaml.Unmarshal([]byte(data), &n)
		if err != nil {
			c.AddFail(err.Error())
			return
		}
		c.NodeMap[configName] = n
	}
}

// processData runs the actual checks against the list of KeyValues provided in
// the Check configuration and determines the Status (Pass or Fail) and Pass or
// Fail messages of the Check Result.
func (c *YamlBase) processData(configName string) {
	for _, kv := range c.Values {
		kvr, fails, err := c.CheckKeyValue(kv, configName)
		switch kvr {
		case KeyValueError:
			c.AddFail(err.Error())
		case KeyValueNotFound:
			c.AddFail(fmt.Sprintf("[%s] '%s' not found", configName, kv.Key))
		case KeyValueNotEqual:
			c.AddFail(fmt.Sprintf("[%s] '%s' equals '%s', expected '%s'", configName, kv.Key, fails[0], kv.Value))
		case KeyValueDisallowedFound:
			c.AddFail(fmt.Sprintf("[%s] disallowed %s: [%s]", configName, kv.Key, strings.Join(fails, ", ")))
		case KeyValueEqual:
			if kv.IsList {
				c.AddPass(fmt.Sprintf("[%s] no disallowed '%s'", configName, kv.Key))
			} else {
				c.AddPass(fmt.Sprintf("[%s] '%s' equals '%s'", configName, kv.Key, kv.Value))
			}
		}
	}
	if len(c.Result.Failures) != 0 {
		c.Result.Status = config.Fail
	} else {
		c.Result.Status = config.Pass
	}
}

// CheckKeyValue lookups the Yaml data for a specific KeyValue and returns the
// result, actual values and errors.
func (c *YamlBase) CheckKeyValue(kv KeyValue, mapKey string) (KeyValueResult, []string, error) {
	node := c.NodeMap[mapKey]
	foundNodes, err := utils.LookupYamlPath(&node, kv.Key)
	if err != nil {
		return KeyValueError, nil, err
	}

	if len(foundNodes) == 0 && !kv.Optional {
		return KeyValueNotFound, nil, nil
	}

	// Throw an error if we are checking a list but no allow/disallow list provided.
	if len(kv.Allowed) == 0 && len(kv.Disallowed) == 0 && kv.IsList {
		return KeyValueError, nil, errors.New("list of allowed or disallowed values not provided")
	}

	// Perform direct comparison if no allow/disallow list provided.
	if len(kv.Allowed) == 0 && len(kv.Disallowed) == 0 {
		notEquals := []string{}
		for _, item := range foundNodes {
			if !kv.Equals(item.Value) && !utils.StringSliceContains(notEquals, item.Value) {
				notEquals = append(notEquals, item.Value)
			}
		}
		if len(notEquals) > 0 {
			return KeyValueNotEqual, notEquals, nil
		}
		return KeyValueEqual, nil, nil
	}

	// Check each yaml value against the disallowed list.
	fails := []string{}
	for _, item := range foundNodes {
		if kv.IsList {
			for _, v := range item.Content {
				if kv.IsDisallowed(v.Value) && !utils.StringSliceContains(fails, v.Value) {
					fails = append(fails, v.Value)
				}
			}
		} else {
			if kv.IsDisallowed(item.Value) && !utils.StringSliceContains(fails, item.Value) {
				fails = append(fails, item.Value)
			}
		}
	}
	if len(fails) > 0 {
		return KeyValueDisallowedFound, fails, nil
	}
	return KeyValueEqual, nil, nil
}
