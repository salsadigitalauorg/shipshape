package drupal_test

import (
	"salsadigitalauorg/shipshape/internal"
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"
	"testing"
)

func mockCheck(configName string) core.YamlBase {
	return core.YamlBase{
		CheckBase: core.CheckBase{
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
	c := mockCheck("core.extension.yml")
	required := []string{
		"node&foo",
		"block",
	}
	disallowed := []string{
		"views_ui",
		"field_ui&bar",
	}
	c.UnmarshalDataMap()
	drupal.CheckModulesInYaml(&c, drupal.FileModule, "core.extension.yml", required, disallowed)
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
	c = mockCheck("core.extension.yml")
	c.DataMap = map[string][]byte{
		"core.extension.yml": []byte(`
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
	drupal.CheckModulesInYaml(&c, drupal.FileModule, "core.extension.yml", required, disallowed)
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

	c = mockCheck("core.extension.yml")
	required = []string{
		"node",
		"block",
	}
	disallowed = []string{
		"views_ui",
		"field_ui",
	}
	c.UnmarshalDataMap()
	drupal.CheckModulesInYaml(&c, drupal.FileModule, "core.extension.yml", required, disallowed)
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
		YamlCheck: core.YamlCheck{
			YamlBase: mockCheck("core.extension.yml"),
		},
		Required:   []string{"node", "block"},
		Disallowed: []string{"views_ui", "field_ui"},
	}
	c.Init("", drupal.FileModule)
	if c.File != "core.extension.yml" {
		t.Errorf("File should be 'core.extension.yml', got %s", c.File)
	}
	if c.IgnoreMissing != true {
		t.Errorf("IgnoreMissing should be 'true', got %t", c.IgnoreMissing)
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
