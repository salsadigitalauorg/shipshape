package drupal_test

import (
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"
	"testing"
)

func TestDrupalConfig(t *testing.T) {
	c := drupal.DrupalConfigBase{}
	c.RunCheck()
	if c.Result.Failures[0] != "no data to run check on" {
		t.Errorf("Check should fail with error 'no data to run check on', got '%+v'", c.Result.Failures[0])
	}

	mockCheck := func() drupal.DrupalConfigBase {
		return drupal.DrupalConfigBase{
			CheckBase: core.CheckBase{
				DataMap: map[string][]byte{
					"data": []byte(`
check:
  interval_days: 7
notification:
  emails:
    - admin@example.com
`),
				},
			},
			YamlCheck: core.YamlCheck{
				Values: []core.KeyValue{
					{
						Key:   "check.interval_days",
						Value: "7",
					},
				},
			},
		}
	}

	c = mockCheck()

	c.RunCheck()
	if c.Result.Status == core.Fail {
		t.Errorf("Check status should be Pass, got %s", c.Result.Status)
	}
	if len(c.Result.Passes) != 1 ||
		c.Result.Passes[0] != "[data] 'check.interval_days' equals '7'" {
		t.Errorf("There should be only 1 Pass and it should be equal to ([data] 'check.interval_days' equals '7'), got %+v", c.Result.Passes)
	}
	if len(c.Result.Failures) != 0 {
		t.Errorf("There should be no Failure, got %+v", c.Result.Failures)
	}

	// Wrong key, correct value.
	c = mockCheck()
	c.Values = []core.KeyValue{
		{
			Key:   "check.interval",
			Value: "7",
		},
	}
	c.RunCheck()
	if c.Result.Status == core.Pass {
		t.Errorf("Check status should be Fail, got %s", c.Result.Status)
	}
	if len(c.Result.Failures) != 1 ||
		c.Result.Failures[0] != "[data] 'check.interval' not found" {
		t.Errorf("There should be only 1 Failure and it should be equal to ([data] 'check.interval' not found), got %+v", c.Result.Failures)
	}
	if len(c.Result.Passes) != 0 {
		t.Errorf("There should be no Pass, got %+v", c.Result.Passes)
	}

	// Correct key, wrong value.
	c = mockCheck()
	c.Values = []core.KeyValue{
		{
			Key:   "check.interval_days",
			Value: "8",
		},
	}
	c.RunCheck()
	if c.Result.Status == core.Pass {
		t.Errorf("Check status should be Fail, got %s", c.Result.Status)
	}
	if len(c.Result.Failures) != 1 ||
		c.Result.Failures[0] != "[data] 'check.interval_days' equals '7'" {
		t.Errorf("There should be only 1 Failure and it should be equal to ([data] 'check.interval_days' equals '7'), got %+v", c.Result.Failures)
	}
	if len(c.Result.Passes) != 0 {
		t.Errorf("There should be no Pass, got %+v", c.Result.Passes)
	}

	// Multiple config values - all correct.
	c = mockCheck()
	c.Values = []core.KeyValue{
		{
			Key:   "check.interval_days",
			Value: "7",
		},
		{
			Key:   "notification.emails[0]",
			Value: "admin@example.com",
		},
	}
	c.RunCheck()
	if c.Result.Status == core.Fail {
		t.Errorf("Check status should be Pass, got %s", c.Result.Status)
	}
	if len(c.Result.Passes) != 2 ||
		c.Result.Passes[0] != "[data] 'check.interval_days' equals '7'" ||
		c.Result.Passes[1] != "[data] 'notification.emails[0]' equals 'admin@example.com'" {
		t.Errorf("There should be 2 Passes, got %+v", c.Result.Passes)
	}
	if len(c.Result.Failures) != 0 {
		t.Errorf("There should be no Failure, got %+v", c.Result.Failures)
	}
}

func TestDrupalFileConfig(t *testing.T) {
	c := drupal.DrupalFileConfigCheck{
		DrupalConfigBase: drupal.DrupalConfigBase{
			YamlCheck: core.YamlCheck{
				Values: []core.KeyValue{
					{
						Key:   "check.interval_days",
						Value: "7",
					},
				},
			},
			ConfigName: "update.settings",
		},
		ConfigPath: "testdata/drupal-file-config",
	}
	c.FetchData()
	if len(c.Result.Failures) > 0 {
		t.Errorf("FetchData should succeed, but failed: %+v", c.Result.Failures)
	}
	if c.DataMap == nil {
		t.Errorf("c.DataMap should be filled, but is empty.")
	}
	c.RunCheck()
	if len(c.Result.Failures) > 0 {
		t.Errorf("RunCheck should succeed, but failed: %+v", c.Result.Failures)
	}
	if c.Result.Status != core.Pass {
		t.Errorf("Check result should be Pass, but got: %s", c.Result.Status)
	}
	if len(c.Result.Passes) != 1 || c.Result.Passes[0] != "[update.settings.yml] 'check.interval_days' equals '7'" {
		t.Errorf("There should be 1 Pass with value \"[update.settings.yml] 'check.interval_days' equals '7'\", but got: %+v", c.Result.Passes)
	}
}

func TestDrupalModules(t *testing.T) {
	c := drupal.DrupalFileModuleCheck{
		DrupalFileConfigCheck: drupal.DrupalFileConfigCheck{
			DrupalConfigBase: drupal.DrupalConfigBase{
				ConfigName: "core.extension",
			},
			ConfigPath: "testdata/drupal-file-config",
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
	c.FetchData()
	if len(c.Result.Failures) > 0 {
		t.Errorf("FetchData should succeed, but failed: %+v", c.Result.Failures)
	}
	if c.DataMap == nil {
		t.Errorf("c.DataMap should be filled, but is empty.")
	}
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
