package drupal_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	yamlv3 "gopkg.in/yaml.v3"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestRegisterChecks(t *testing.T) {
	checksMap := map[config.CheckType]string{
		DrushYaml:     "*drupal.DrushYamlCheck",
		FileModule:    "*drupal.FileModuleCheck",
		DbModule:      "*drupal.DbModuleCheck",
		DbPermissions: "*drupal.DbPermissionsCheck",
		TrackingCode:  "*drupal.TrackingCodeCheck",
		UserRole:      "*drupal.UserRoleCheck",
		AdminUser:     "*drupal.AdminUserCheck",
		DbUserTfa:     "*drupal.DbUserTfaCheck",
	}
	for ct, ts := range checksMap {
		c := config.ChecksRegistry[ct]()
		ctype := reflect.TypeOf(c).String()
		assert.Equal(t, ts, ctype)
	}
}

func TestModuleYamlKey(t *testing.T) {
	assert := assert.New(t)

	kv := ModuleYamlKey(FileModule, "foo")
	assert.Equal(kv.Key, "module.foo")
	assert.Equal(kv.Value, "0")

	kv = ModuleYamlKey(DbModule, "bar")
	assert.Equal(kv.Key, "bar.status")
	assert.Equal(kv.Value, "Enabled")
}

func TestDetermineModuleStatus(t *testing.T) {
	tests := []struct {
		name             string
		yaml             string
		ct               config.CheckType
		modules          []string
		expectedEnabled  []string
		expectedErrored  []string
		expectedDisabled []string
	}{
		{
			name: "all enabled - file",
			yaml: `
module:
  clamav: 0
  tfa: 1
`,
			modules:          []string{"clamav", "tfa"},
			ct:               FileModule,
			expectedEnabled:  []string{"clamav", "tfa"},
			expectedErrored:  []string{},
			expectedDisabled: []string{},
		},
		{
			name: "all enabled - db",
			yaml: `
clamav:
  status: Enabled
tfa:
  status: Enabled
`,
			modules:          []string{"clamav", "tfa"},
			ct:               DbModule,
			expectedEnabled:  []string{"clamav", "tfa"},
			expectedErrored:  []string{},
			expectedDisabled: []string{},
		},
		{
			name: "all disabled - file",
			yaml: `
module:
  some: 0
  other: 1
`,
			modules:          []string{"clamav", "tfa"},
			ct:               DbModule,
			expectedEnabled:  []string{},
			expectedErrored:  []string{},
			expectedDisabled: []string{"clamav", "tfa"},
		},
		{
			name: "all disabled - db",
			yaml: `
clamav:
  status: Disabled
tfa:
  status: Disabled
`,
			modules:          []string{"clamav", "tfa"},
			ct:               DbModule,
			expectedEnabled:  []string{},
			expectedErrored:  []string{},
			expectedDisabled: []string{"clamav", "tfa"},
		},
		{
			name: "some enabled - file",
			yaml: `
module:
  clamav: 0
  some: 1
`,
			modules:          []string{"clamav", "tfa"},
			ct:               FileModule,
			expectedEnabled:  []string{"clamav"},
			expectedErrored:  []string{},
			expectedDisabled: []string{"tfa"},
		},
		{
			name: "some enabled - db",
			yaml: `
clamav:
  status: Enabled
tfa:
  status: Disabled
`,
			modules:          []string{"clamav", "tfa"},
			ct:               DbModule,
			expectedEnabled:  []string{"clamav"},
			expectedErrored:  []string{},
			expectedDisabled: []string{"tfa"},
		},
		{
			name: "some errored - file",
			yaml: `
module:
  clamav: 0
  tfa: 1
`,
			modules:          []string{"clamav", "tfa&>"},
			ct:               FileModule,
			expectedEnabled:  []string{"clamav"},
			expectedErrored:  []string{"invalid character '&' at position 10, following \".tfa\""},
			expectedDisabled: []string{},
		},
		{
			name: "some errored - db",
			yaml: `
clamav:
  status: Enabled
tfa:
  status: Disabled
`,
			modules:          []string{"clamav", "tfa&>"},
			ct:               DbModule,
			expectedEnabled:  []string{"clamav"},
			expectedErrored:  []string{"invalid character '&' at position 3, following \"tfa\""},
			expectedDisabled: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := yamlv3.Node{}
			yamlData := []byte(tt.yaml)
			err := yamlv3.Unmarshal(yamlData, &n)
			if err != nil {
				t.Error(err)
			}

			assert := assert.New(t)
			enabled, errored, disabled := DetermineModuleStatus(n, tt.ct, tt.modules)
			assert.ElementsMatch(enabled, tt.expectedEnabled, "expected enabled modules to match")
			assert.ElementsMatch(errored, tt.expectedErrored, "expected errors to match")
			assert.ElementsMatch(disabled, tt.expectedDisabled, "expected disabled modules to match")
		})
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
