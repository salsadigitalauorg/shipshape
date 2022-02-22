package drupal_test

import (
	"reflect"
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
	if c.Result.Status != core.Fail {
		t.Error("CheckModulesInYaml should Fail")
	}
	expectedFails := []string{
		"invalid character '&' at position 11, following \".node\"",
		"invalid character '&' at position 15, following \".field_ui\"",
	}
	if len(c.Result.Failures) != 2 || !reflect.DeepEqual(expectedFails, c.Result.Failures) {
		t.Errorf("There should be exactly 2 Failures, got: %#v", c.Result.Failures)
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
	if c.Result.Status != core.Fail {
		t.Error("Check should fail")
	}
	expectedFails = []string{
		"'node' is not enabled",
		"'views_ui' is enabled",
	}
	if len(c.Result.Failures) != 2 || !reflect.DeepEqual(expectedFails, c.Result.Failures) {
		t.Errorf("There should be exactly 2 Failures, got %#v", c.Result.Failures)
	}
	expectedPasses := []string{
		"'block' is enabled",
		"'field_ui' is not enabled",
	}
	if len(c.Result.Passes) != 2 || !reflect.DeepEqual(expectedPasses, c.Result.Passes) {
		t.Errorf("There should be exactly 2 Passes, got %#v", c.Result.Passes)
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
	if c.Result.Status != core.Pass {
		t.Error("Check should Pass")
	}
	if len(c.Result.Failures) > 0 {
		t.Errorf("RunCheck should succeed, but failed: %+v", c.Result.Failures)
	}
	expectedPasses = []string{
		"'node' is enabled",
		"'block' is enabled",
		"'views_ui' is not enabled",
		"'field_ui' is not enabled",
	}
	if len(c.Result.Passes) != 4 || !reflect.DeepEqual(expectedPasses, c.Result.Passes) {
		t.Errorf("There should be 4 Passes, but got: %+v", len(c.Result.Passes))
	}
	if len(c.Result.Failures) != 0 {
		t.Errorf("There should be 0 Failures, but got: %+v", c.Result.Failures)
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
	c.UnmarshalDataMap()
	c.RunCheck()
}
