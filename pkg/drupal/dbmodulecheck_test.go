package drupal_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestDbModuleMerge(t *testing.T) {
	assert := assert.New(t)

	c := drupal.DbModuleCheck{
		DrushYamlCheck: drupal.DrushYamlCheck{
			YamlBase: shipshape.YamlBase{
				Values: []shipshape.KeyValue{
					{Key: "key1", Value: "val1", Optional: false},
				},
			},
		},
		Required:   []string{"req1"},
		Disallowed: []string{"disallowed1"},
	}
	c.Merge(&drupal.DbModuleCheck{
		DrushYamlCheck: drupal.DrushYamlCheck{
			YamlBase: shipshape.YamlBase{
				Values: []shipshape.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Required:   []string{"req2"},
		Disallowed: []string{"disallowed2"},
	})
	assert.EqualValues(drupal.DbModuleCheck{
		DrushYamlCheck: drupal.DrushYamlCheck{
			YamlBase: shipshape.YamlBase{
				Values: []shipshape.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Required:   []string{"req1", "req2"},
		Disallowed: []string{"disallowed1", "disallowed2"},
	}, c)
}

func TestDbModuleCheck(t *testing.T) {
	assert := assert.New(t)

	c := drupal.DbModuleCheck{}
	c.Init(drupal.DbModule)
	assert.Equal("pm:list --status=enabled", c.Command)

	mockCheck := func(dataMap map[string][]byte) drupal.DbModuleCheck {
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
		c := drupal.DbModuleCheck{
			DrushYamlCheck: drupal.DrushYamlCheck{
				YamlBase: shipshape.YamlBase{
					CheckBase: shipshape.CheckBase{DataMap: dataMap},
				},
				ConfigName: "modules",
			},
			Required:   []string{"block", "node"},
			Disallowed: []string{"views_ui", "field_ui"},
		}
		c.Init(drupal.DbModule)
		c.UnmarshalDataMap()
		c.RunCheck()
		return c
	}

	c = mockCheck(nil)
	assert.Equal(shipshape.Pass, c.Result.Status)
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

	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'node' is enabled",
		"'field_ui' is not enabled",
	})
	assert.ElementsMatch(c.Result.Failures, []string{
		"'block' is not enabled",
		"'views_ui' is enabled",
	})
}
