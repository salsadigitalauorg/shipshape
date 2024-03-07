//go:generate go run ../../../cmd/gen.go registry --checkpackage=json

package json

import (
	"errors"
	"fmt"

	"github.com/goccy/go-json"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// UnmarshalDataMap parses the DataMap into Json for further processing.
// DataMap is expected to be populated from FetchData in the respective Check
// implementation.
func (c *JsonCheck) UnmarshalDataMap() {
	c.Node = map[string]any{}
	for configName, data := range c.DataMap {
		var n any
		err := json.Unmarshal(data, &n)
		if err != nil {
			c.AddBreach(&breach.ValueBreach{ValueLabel: "JSON error", Value: err.Error()})
			return
		}
		c.Node[configName] = n
	}
}

// processData runs the actual checks against the list of KeyValues provided in
// the Check configuration and determines the Status (Pass or Fail) and Pass or
// Fail messages of the Check Result.
func (c *JsonCheck) processData(configName string) {
	for _, kv := range c.KeyValues {
		kvr, fails, err := CheckKeyValue(c.Node[configName], kv)
		switch kvr {
		case yaml.KeyValueError:
			c.AddBreach(&breach.ValueBreach{Value: err.Error()})
		case yaml.KeyValueNotFound:
			c.AddBreach(&breach.KeyValueBreach{
				KeyLabel:   "config",
				Key:        configName,
				ValueLabel: "key not found",
				Value:      kv.Key,
			})
		case yaml.KeyValueNotEqual:
			c.AddBreach(&breach.KeyValueBreach{
				KeyLabel:      configName,
				Key:           kv.Key,
				ValueLabel:    "actual",
				ExpectedValue: kv.Value,
				Value:         fails[0],
			})
		case yaml.KeyValueDisallowedFound:
			c.AddBreach(&breach.KeyValuesBreach{
				KeyLabel:   "config",
				Key:        configName,
				ValueLabel: fmt.Sprintf("disallowed %s", kv.Key),
				Values:     fails,
			})
		case yaml.KeyValueEqual:
			if kv.IsList {
				c.AddPass(fmt.Sprintf("[%s] no disallowed '%s'", configName, kv.Key))
			} else {
				c.AddPass(fmt.Sprintf("[%s] '%s' equals '%s'", configName, kv.Key, kv.Value))
			}
		}
	}
	if len(c.Result.Breaches) != 0 {
		c.Result.Status = result.Fail
	} else {
		c.Result.Status = result.Pass
	}
}

// CheckKeyValue lookups the Json data for a specific KeyValue and returns the
// result, actual values and errors.
func CheckKeyValue(node any, kv KeyValue) (yaml.KeyValueResult, []string, error) {
	foundValues, err, _ := LookupJson(kv.Key, node)
	if err != nil {
		return yaml.KeyValueError, nil, err
	}

	if foundValues == nil {
		if !kv.Optional {
			return yaml.KeyValueNotFound, nil, nil
		}
		return yaml.KeyValueEqual, nil, nil
	}

	// Throw an error if we are checking a list but no allow/disallow list provided.
	if len(kv.AllowedValues) == 0 && len(kv.DisallowedValues) == 0 && kv.IsList {
		return yaml.KeyValueError, nil, errors.New("list of allowed or disallowed values not provided")
	}

	var foundNodes []any
	switch foundValues.(type) {
	case []any:
		if kv.IsList {
			foundNodes = foundValues.([]any)
		} else {
			return yaml.KeyValueError, nil, errors.New("A list of values was found but is-list is not set")
		}
	default:
		foundNodes = []any{foundValues}
	}

	// Perform direct comparison if no allow/disallow list provided.
	if len(kv.AllowedValues) == 0 && len(kv.DisallowedValues) == 0 {
		var notEquals []string
		for _, item := range foundNodes {
			if !kv.Equals(item) && !utils.SliceContains(notEquals, item) {
				notEquals = append(notEquals, fmt.Sprint(item))
			}
		}
		if len(notEquals) > 0 {
			return yaml.KeyValueNotEqual, notEquals, nil
		}
		return yaml.KeyValueEqual, nil, nil
	}

	// Check each yaml value against the disallowed list.
	var fails []string
	for _, item := range foundNodes {
		if kv.IsDisallowed(item) && !utils.SliceContains(fails, item) {
			fails = append(fails, fmt.Sprint(item))
		}
	}
	if len(fails) > 0 {
		return yaml.KeyValueDisallowedFound, fails, nil
	}
	return yaml.KeyValueEqual, nil, nil
}
