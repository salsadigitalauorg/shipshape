package drupal_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"

	"github.com/stretchr/testify/assert"
)

func TestDbModuleMerge(t *testing.T) {
	assert := assert.New(t)

	c := DbModuleCheck{
		DrushYamlCheck: DrushYamlCheck{
			YamlBase: yaml.YamlBase{
				Values: []yaml.KeyValue{
					{Key: "key1", Value: "val1", Optional: false},
				},
			},
		},
		Required:   []string{"req1"},
		Disallowed: []string{"disallowed1"},
	}
	c.Merge(&DbModuleCheck{
		DrushYamlCheck: DrushYamlCheck{
			YamlBase: yaml.YamlBase{
				Values: []yaml.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Required:   []string{"req2"},
		Disallowed: []string{"disallowed2"},
	})
	assert.EqualValues(DbModuleCheck{
		DrushYamlCheck: DrushYamlCheck{
			YamlBase: yaml.YamlBase{
				Values: []yaml.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Required:   []string{"req2"},
		Disallowed: []string{"disallowed2"},
	}, c)
}

func TestDbModuleCheck(t *testing.T) {
	assert := assert.New(t)

	c := DbModuleCheck{}
	c.Init(DbModule)
	assert.Equal("pm:list --status=enabled", c.Command)

	mockCheck := func(dataMap map[string][]byte) DbModuleCheck {
		if dataMap == nil {
			dataMap = map[string][]byte{
				"modules": []byte(`
block:
  status: Enabled
node:
  status: Enabled

`),
			}
		}
		c := DbModuleCheck{
			DrushYamlCheck: DrushYamlCheck{
				YamlBase: yaml.YamlBase{
					CheckBase: config.CheckBase{DataMap: dataMap},
				},
				ConfigName: "modules",
			},
			Required:   []string{"block", "node"},
			Disallowed: []string{"views_ui", "field_ui"},
		}
		c.Init(DbModule)
		c.UnmarshalDataMap()
		c.RunCheck()
		return c
	}

	c = mockCheck(nil)
	assert.Equal(result.Pass, c.Result.Status)
	assert.Empty(c.Result.Breaches)
	assert.ElementsMatch(c.Result.Passes, []string{
		"all required modules are enabled",
		"all disallowed modules are disabled",
	})

	c = mockCheck(map[string][]byte{
		"modules": []byte(`
node:
  status: Enabled
views_ui:
  status: Enabled

`),
	})

	assert.Equal(result.Fail, c.Result.Status)
	assert.ElementsMatch(c.Result.Passes, []string{
		"some required modules are enabled: node",
		"some disallowed modules are disabled: field_ui",
	})
	assert.ElementsMatch(
		[]result.Breach{
			&result.KeyValuesBreach{
				BreachType: "key-values",
				CheckType:  "drupal-db-module",
				Severity:   "normal",
				Key:        "required modules are not enabled",
				Values:     []string{"block"},
			},
			&result.KeyValuesBreach{
				BreachType: "key-values",
				CheckType:  "drupal-db-module",
				Severity:   "normal",
				Key:        "disallowed modules are enabled",
				Values:     []string{"views_ui"},
			},
		},
		c.Result.Breaches,
	)
}
