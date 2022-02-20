package drupal_test

import (
	"reflect"
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"
	"testing"
)

func TestDrupalFileModule(t *testing.T) {
	mockCheck := func() drupal.FileModuleCheck {
		return drupal.FileModuleCheck{
			YamlCheck: core.YamlCheck{
				YamlBase: core.YamlBase{
					CheckBase: core.CheckBase{
						DataMap: map[string][]byte{
							"core.extension.yml": []byte(`
module:
  block: 0
  node: 0

`),
						},
					},
				},
				File: "core.extension",
			},
		}
	}

	// Invalid yaml key.
	c := mockCheck()
	c.Required = []string{
		"node&foo",
		"block",
	}
	c.Disallowed = []string{
		"views_ui",
		"field_ui&bar",
	}
	c.File = ""
	c.Init("", drupal.FileModule)
	if c.File != "core.extension" {
		t.Errorf("File should be 'core.extension', got %s", c.File)
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if c.Result.Status != core.Fail {
		t.Error("RunCheck should Fail")
	}
	expectedFails := []string{
		"invalid character '&' at position 11, following \".node\"",
		"invalid character '&' at position 15, following \".field_ui\"",
	}
	if len(c.Result.Failures) != 2 || !reflect.DeepEqual(expectedFails, c.Result.Failures) {
		t.Errorf("There should be exactly 2 Failures, got: %#v", c.Result.Failures)
	}

	// Required is not enabled & disallowed is enabled.
	c = mockCheck()
	c.DataMap = map[string][]byte{
		"core.extension.yml": []byte(`
module:
  block: 0
  views_ui: 0

`),
	}
	c.Required = []string{
		"node",
		"block",
	}
	c.Disallowed = []string{
		"views_ui",
		"field_ui",
	}
	c.UnmarshalDataMap()
	c.RunCheck()
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

	c = mockCheck()
	c.Required = []string{
		"node",
		"block",
	}
	c.Disallowed = []string{
		"views_ui",
		"field_ui",
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if len(c.Result.Failures) > 0 {
		t.Errorf("RunCheck should succeed, but failed: %+v", c.Result.Failures)
	}
	if c.Result.Status != core.Pass {
		t.Errorf("Check result should be Pass, but got: %s", c.Result.Status)
	}
	if len(c.Result.Passes) != 4 {
		t.Errorf("There should be 4 Passes, but got: %+v", len(c.Result.Passes))
	}
	if len(c.Result.Failures) != 0 {
		t.Errorf("There should be 0 Failures, but got: %+v", c.Result.Failures)
	}
	if c.Result.Passes[0] != "'node' is enabled" ||
		c.Result.Passes[1] != "'block' is enabled" ||
		c.Result.Passes[2] != "'views_ui' is not enabled" ||
		c.Result.Passes[3] != "'field_ui' is not enabled" {
		t.Errorf("Wrong pass messages, got: %+v", c.Result.Passes)
	}
}
