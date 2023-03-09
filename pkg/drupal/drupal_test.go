package drupal_test

import (
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	. "github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"

	"github.com/stretchr/testify/assert"
)

func TestRegisterChecks(t *testing.T) {
	checksMap := map[config.CheckType]string{
		DrushYaml:     "*drupal.DrushYamlCheck",
		FileModule:    "*drupal.FileModuleCheck",
		DbModule:      "*drupal.DbModuleCheck",
		DbPermissions: "*drupal.DbPermissionsCheck",
		TrackingCode:  "*drupal.TrackingCodeCheck",
		UserRole:      "*drupal.UserRoleCheck",
	}
	for ct, ts := range checksMap {
		c := config.ChecksRegistry[ct]()
		ctype := reflect.TypeOf(c).String()
		assert.Equal(t, ts, ctype)
	}
}

func mockCheck(configName string) shipshape.YamlBase {
	return shipshape.YamlBase{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				configName: []byte(`
module:
  block: 0
  node: 0

`),
			},
		},
	}
}

func TestCheckModulesInYamlDisallowedIsEnabled(t *testing.T) {
	assert := assert.New(t)

	c := mockCheck("shipshape.extension.yml")
	c.DataMap = map[string][]byte{
		"shipshape.extension.yml": []byte(`
module:
  clamav: 0
  tfa: 0
  dblog: 0

`),
	}
	c.UnmarshalDataMap()

	required := []string{
		"clamav",
		"tfa",
	}
	disallowed := []string{
		"dblog",
		"module_permissions_ui",
		"update",
	}

	CheckModulesInYaml(&c, FileModule, "shipshape.extension.yml", required, disallowed)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'clamav' is enabled",
		"'tfa' is enabled",
		"'module_permissions_ui' is not enabled",
		"'update' is not enabled",
	})
	assert.ElementsMatch(c.Result.Failures, []string{"'dblog' is enabled"})
}

func TestCheckModulesInYaml(t *testing.T) {
	assert := assert.New(t)

	// Invalid yaml key.
	c := mockCheck("shipshape.extension.yml")
	required := []string{
		"node&foo",
		"block",
	}
	disallowed := []string{
		"views_ui",
		"field_ui&bar",
	}
	c.UnmarshalDataMap()
	CheckModulesInYaml(&c, FileModule, "shipshape.extension.yml", required, disallowed)
	assert.Equal(config.Fail, c.Result.Status)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'block' is enabled",
		"'views_ui' is not enabled",
	})
	assert.ElementsMatch(c.Result.Failures, []string{
		"invalid character '&' at position 11, following \".node\"",
		"invalid character '&' at position 15, following \".field_ui\"",
	})

	// Required is not enabled & disallowed is enabled.
	c = mockCheck("shipshape.extension.yml")
	c.DataMap = map[string][]byte{
		"shipshape.extension.yml": []byte(`
module:
  block: 0
  views_ui: 0

`),
	}
	required = []string{
		"node",
		"block",
	}
	disallowed = []string{
		"views_ui",
		"field_ui",
	}
	c.UnmarshalDataMap()
	CheckModulesInYaml(&c, FileModule, "shipshape.extension.yml", required, disallowed)
	assert.Equal(config.Fail, c.Result.Status)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'block' is enabled",
		"'field_ui' is not enabled",
	})
	assert.ElementsMatch(c.Result.Failures, []string{
		"'node' is not enabled",
		"'views_ui' is enabled",
	})

	c = mockCheck("shipshape.extension.yml")
	required = []string{
		"node",
		"block",
	}
	disallowed = []string{
		"views_ui",
		"field_ui",
	}
	c.UnmarshalDataMap()
	CheckModulesInYaml(&c, FileModule, "shipshape.extension.yml", required, disallowed)
	assert.Equal(config.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch(c.Result.Passes, []string{
		"'node' is enabled",
		"'block' is enabled",
		"'views_ui' is not enabled",
		"'field_ui' is not enabled",
	})
}
