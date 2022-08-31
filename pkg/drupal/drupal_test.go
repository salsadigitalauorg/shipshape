package drupal_test

import (
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func TestRegisterChecks(t *testing.T) {
	checksMap := map[shipshape.CheckType]string{
		drupal.DrushYaml:     "*drupal.DrushYamlCheck",
		drupal.FileModule:    "*drupal.FileModuleCheck",
		drupal.DbModule:      "*drupal.DbModuleCheck",
		drupal.DbPermissions: "*drupal.DbPermissionsCheck",
		drupal.TrackingCode:  "*drupal.TrackingCodeCheck",
		drupal.UserRole:      "*drupal.UserRoleCheck",
	}
	for ct, ts := range checksMap {
		c := shipshape.ChecksRegistry[ct]()
		ctype := reflect.TypeOf(c).String()
		if ctype != ts {
			t.Errorf("expecting check of type '%s', got '%s'", ts, ctype)
		}
	}
}

func mockCheck(configName string) shipshape.YamlBase {
	return shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
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

func TestCheckModulesInYaml(t *testing.T) {
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
	drupal.CheckModulesInYaml(&c, drupal.FileModule, "shipshape.extension.yml", required, disallowed)
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"'block' is enabled",
		"'views_ui' is not enabled",
	}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{
		"invalid character '&' at position 11, following \".node\"",
		"invalid character '&' at position 15, following \".field_ui\"",
	}); !ok {
		t.Error(msg)
	}

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
	drupal.CheckModulesInYaml(&c, drupal.FileModule, "shipshape.extension.yml", required, disallowed)
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"'block' is enabled",
		"'field_ui' is not enabled",
	}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{
		"'node' is not enabled",
		"'views_ui' is enabled",
	}); !ok {
		t.Error(msg)
	}

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
	drupal.CheckModulesInYaml(&c, drupal.FileModule, "shipshape.extension.yml", required, disallowed)
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"'node' is enabled",
		"'block' is enabled",
		"'views_ui' is not enabled",
		"'field_ui' is not enabled",
	}); !ok {
		t.Error(msg)
	}
}

func TestFileModuleCheck(t *testing.T) {
	c := drupal.FileModuleCheck{
		YamlCheck: shipshape.YamlCheck{
			YamlBase: mockCheck("core.extension.yml"),
		},
		Required:   []string{"node", "block"},
		Disallowed: []string{"views_ui", "field_ui"},
	}
	c.Init(drupal.FileModule)
	if c.File != "core.extension.yml" {
		t.Errorf("File should be 'core.extension.yml', got %s", c.File)
	}
	if *c.IgnoreMissing != true {
		t.Errorf("IgnoreMissing should be 'true', got %t", *c.IgnoreMissing)
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"'node' is enabled",
		"'block' is enabled",
		"'views_ui' is not enabled",
		"'field_ui' is not enabled",
	}); !ok {
		t.Error(msg)
	}
}
