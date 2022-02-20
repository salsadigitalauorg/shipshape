package drupal_test

import (
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"
	"testing"
)

func TestDrupalFileModule(t *testing.T) {
	c := drupal.FileModuleCheck{
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
		Required: []string{
			"node",
			"block",
		},
		Disallowed: []string{
			"views_ui",
			"field_ui",
		},
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
