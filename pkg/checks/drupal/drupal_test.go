package drupal_test

import (
	"reflect"
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"

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

func mockCheck(configName string) yaml.YamlBase {
	return yaml.YamlBase{
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
		"all required modules are enabled",
		"all disallowed modules are disabled",
	})
	assert.ElementsMatch(c.Result.Failures, []string{"disallowed modules are enabled: dblog"})
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
	assert.Equal(result.Fail, c.Result.Status)
	assert.ElementsMatch(c.Result.Passes, []string{
		"some required modules are enabled: block",
		"some disallowed modules are disabled: views_ui",
	})
	assert.ElementsMatch(c.Result.Failures, []string{
		"error verifying status for required modules: invalid character '&' at position 11, following \".node\"",
		"error verifying status for disallowed modules: invalid character '&' at position 15, following \".field_ui\"",
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
	assert.Equal(result.Fail, c.Result.Status)
	assert.ElementsMatch(c.Result.Passes, []string{
		"some required modules are enabled: block",
		"some disallowed modules are disabled: field_ui",
	})
	assert.ElementsMatch(c.Result.Failures, []string{
		"required modules are not enabled: node",
		"disallowed modules are enabled: views_ui",
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
	assert.Equal(result.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch(c.Result.Passes, []string{
		"all required modules are enabled",
		"all disallowed modules are disabled",
	})
}
