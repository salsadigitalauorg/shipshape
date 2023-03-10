package drupal_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	. "github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/yaml"

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
  status: enabled
node:
  status: enabled

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
	assert.Equal(config.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'block' is enabled",
		"'node' is enabled",
		"'views_ui' is not enabled",
		"'field_ui' is not enabled",
	})

	c = mockCheck(map[string][]byte{
		"modules": []byte(`
node:
  status: enabled
views_ui:
  status: enabled

`),
	})

	assert.Equal(config.Fail, c.Result.Status)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'node' is enabled",
		"'field_ui' is not enabled",
	})
	assert.ElementsMatch(c.Result.Failures, []string{
		"'block' is not enabled",
		"'views_ui' is enabled",
	})
}
