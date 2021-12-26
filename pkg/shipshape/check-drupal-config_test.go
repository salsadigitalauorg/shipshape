package shipshape_test

import (
	"salsadigitalauorg/shipshape/pkg/shipshape"
	"testing"
)

func TestDrupalConfigRunCheck(t *testing.T) {
	c := shipshape.DrupalConfigBase{}
	err := c.RunCheck()
	if err == nil || err.Error() != "no data to run check on" {
		t.Errorf("Check should fail with error 'no data to run check on', got '%+v'", err)
	}

	c = shipshape.DrupalConfigBase{
		CheckBase: shipshape.CheckBase{
			Data: []byte(`
check:
  interval_days: 7
notification:
  emails:
    - admin@example.com
`),
		},
		YamlCheck: shipshape.YamlCheck{
			Values: []shipshape.KeyValue{
				{
					Key:   "check.interval_days",
					Value: "7",
				},
			},
		},
	}

	err = c.RunCheck()
	if err != nil {
		t.Errorf("Check should be successful, got error: %+v", err)
	}
	if c.Result.Status == shipshape.Fail {
		t.Errorf("Check status should be Pass, got %s", c.Result.Status)
	}
	if len(c.Result.Passes) != 1 ||
		c.Result.Passes[0] != "'check.interval_days' is equal to '7'" {
		t.Errorf("There should be only 1 Pass and it should be equal to ('check.interval_days' is equal to '7'), got %+v", c.Result.Passes)
	}
	if len(c.Result.Failures) != 0 {
		t.Errorf("There should be no Failure, got %+v", c.Result.Failures)
	}

	// Wrong key, correct value.
	c.Values = []shipshape.KeyValue{
		{
			Key:   "check.interval",
			Value: "7",
		},
	}
	err = c.RunCheck()
	if err != nil {
		t.Errorf("Check should be successful, got error: %+v", err)
	}
	if c.Result.Status == shipshape.Pass {
		t.Errorf("Check status should be Fail, got %s", c.Result.Status)
	}
	if len(c.Result.Failures) != 1 ||
		c.Result.Failures[0] != "No value found for 'check.interval'" {
		t.Errorf("There should be only 1 Failure and it should be equal to (No value found for 'check.interval'), got %+v", c.Result.Failures)
	}
	if len(c.Result.Passes) != 0 {
		t.Errorf("There should be no Pass, got %+v", c.Result.Passes)
	}

	// Correct key, wrong value.
	c.Values = []shipshape.KeyValue{
		{
			Key:   "check.interval_days",
			Value: "8",
		},
	}
	err = c.RunCheck()
	if err != nil {
		t.Errorf("Check should be successful, got error: %+v", err)
	}
	if c.Result.Status == shipshape.Pass {
		t.Errorf("Check status should be Fail, got %s", c.Result.Status)
	}
	if len(c.Result.Failures) != 1 ||
		c.Result.Failures[0] != "'check.interval_days' is not equal to '8'" {
		t.Errorf("There should be only 1 Failure and it should be equal to ('check.interval_days' is not equal to '8'), got %+v", c.Result.Failures)
	}
	if len(c.Result.Passes) != 0 {
		t.Errorf("There should be no Pass, got %+v", c.Result.Passes)
	}

	// Multiple config values - all correct.
	c.Values = []shipshape.KeyValue{
		{
			Key:   "check.interval_days",
			Value: "7",
		},
		{
			Key:   "notification.emails[0]",
			Value: "admin@example.com",
		},
	}
	err = c.RunCheck()
	if err != nil {
		t.Errorf("Check should be successful, got error: %+v", err)
	}
	if c.Result.Status == shipshape.Fail {
		t.Errorf("Check status should be Pass, got %s", c.Result.Status)
	}
	if len(c.Result.Passes) != 2 ||
		c.Result.Passes[0] != "'check.interval_days' is equal to '7'" ||
		c.Result.Passes[1] != "'notification.emails[0]' is equal to 'admin@example.com'" {
		t.Errorf("There should be 2 Passes, got %+v", c.Result.Passes)
	}
	if len(c.Result.Failures) != 0 {
		t.Errorf("There should be no Failure, got %+v", c.Result.Failures)
	}
}
