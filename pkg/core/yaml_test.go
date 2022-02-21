package core_test

import (
	"reflect"
	"salsadigitalauorg/shipshape/pkg/core"
	"testing"
)

func TestYamlUnmarshalDataMap(t *testing.T) {
	// Invalid data.
	c := core.YamlBase{
		CheckBase: core.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
	baz
`),
			},
		},
	}
	c.UnmarshalDataMap()
	if c.Result.Status != core.Fail {
		t.Error("invalid yaml data should Fail")
	}
	if len(c.Result.Failures) != 1 || c.Result.Failures[0] != "yaml: line 4: found character that cannot start any token" {
		t.Errorf("there should be exactly 1 error, got %#v", c.Result.Failures)
	}

	// Valid data.
	c = core.YamlBase{
		CheckBase: core.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    - name: baz
      value: zoom
`),
			},
		},
	}
	c.UnmarshalDataMap()
	if len(c.Result.Failures) > 0 {
		t.Errorf("there should be no error, got %#v", c.Result.Failures)
	}

	// Invalid yaml kec.
	c = core.YamlBase{
		CheckBase: core.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    baz&*zoom: zap
`),
			},
		},
		Values: []core.KeyValue{
			{Key: "baz&*zoom", Value: "zap"},
		},
	}
	c.RunCheck()
	if c.Result.Status != core.Fail {
		t.Error("invalid yaml key should Fail")
	}
	if len(c.Result.Failures) != 1 || c.Result.Failures[0] != "invalid character '&' at position 3, following \"baz\"" {
		t.Errorf("there should be exactly 1 error, got %#v", c.Result.Failures)
	}
}

func TestYamlCheckKeyValue(t *testing.T) {
	c := core.YamlBase{
		CheckBase: core.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    - baz: zoo

`),
			},
		},
	}
	c.UnmarshalDataMap()

	// Invalid path.
	_, _, err := c.CheckKeyValue(core.KeyValue{
		Key:   "&*&^);",
		Value: "foo",
	}, "data")
	if err == nil || err.Error() != "child name missing at position 0, following \"\"" {
		t.Error("should fail with error 'child name missing at position 0, following \"\"', got success")
	}

	// Non-existent path.
	kvr, _, err := c.CheckKeyValue(core.KeyValue{
		Key:   "foo.baz",
		Value: "foo",
	}, "data")
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != core.KeyValueNotFound {
		t.Errorf("result should be KeyValueNotFound(-1), got %d", kvr)
	}

	// Wrong value.
	kvr, _, err = c.CheckKeyValue(core.KeyValue{
		Key:   "foo.bar[0].baz",
		Value: "zoom",
	}, "data")
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != core.KeyValueNotEqual {
		t.Errorf("result should be KeyValueNotEqual(0), got %d", kvr)
	}

	// Correct value.
	kvr, _, err = c.CheckKeyValue(core.KeyValue{
		Key:   "foo.bar[0].baz",
		Value: "zoo",
	}, "data")
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != core.KeyValueEqual {
		t.Errorf("result should be KeyValueEqual(1), got %d", kvr)
	}
}

func TestYamlCheckKeyValueList(t *testing.T) {
	c := core.YamlBase{
		CheckBase: core.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    - baz
    - zoo
    - zoom

`),
			},
		},
	}
	c.UnmarshalDataMap()

	// Disallowed list not provided.
	_, _, err := c.CheckKeyValue(core.KeyValue{
		Key:    "foo.bar",
		IsList: true,
	}, "data")
	if err == nil {
		t.Error("should fail")
	}

	var kvr core.KeyValueResult
	var fails []string
	// Disallowed values in yaml.
	kvr, fails, err = c.CheckKeyValue(core.KeyValue{
		Key:        "foo.bar",
		IsList:     true,
		Disallowed: []string{"baz", "zoo"},
	}, "data")
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != core.KeyValueDisallowedFound {
		t.Errorf("result should be KeyValueDisallowedFound(-2), got %d", kvr)
	}
	expectedFails := []string{"baz", "zoo"}
	if len(fails) != 2 || !reflect.DeepEqual(fails, expectedFails) {
		t.Errorf("There should be exactly 2 Failures, with values %+v, got %+v", expectedFails, fails)
	}

	// No disallowed values in yaml.
	kvr, fails, _ = c.CheckKeyValue(core.KeyValue{
		Key:        "foo.bar",
		IsList:     true,
		Disallowed: []string{"this should", "be a success"},
	}, "data")
	if kvr != core.KeyValueEqual {
		t.Errorf("result should be KeyValueEqual(1), got %d", kvr)
	}
	if len(fails) > 0 {
		t.Errorf("There should be no Failures, got %+v", fails)
	}

}

func TestYamlBase(t *testing.T) {
	c := core.YamlBase{}
	c.HasData(true)
	if c.Result.Failures[0] != "no data available" {
		t.Errorf("Check should fail with error 'no data available', got '%+v'", c.Result.Failures[0])
	}

	mockCheck := func() core.YamlBase {
		return core.YamlBase{
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
			Values: []core.KeyValue{
				{
					Key:   "check.interval_days",
					Value: "7",
				},
			},
		}
	}

	c = mockCheck()
	c.UnmarshalDataMap()
	c.RunCheck()
	if c.Result.Status == core.Fail {
		t.Logf("result: %#v", c.Result)
		t.Error("Check should Pass")
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
	c.UnmarshalDataMap()
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
	c.UnmarshalDataMap()
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

func TestYamlBaseListValues(t *testing.T) {
	mockCheck := func() core.YamlBase {
		return core.YamlBase{
			CheckBase: core.CheckBase{
				DataMap: map[string][]byte{
					"data": []byte(`
foo:
  - a
  - b
  - c
  - d
`),
				},
			},
			Values: []core.KeyValue{
				{
					Key:        "foo",
					IsList:     true,
					Disallowed: []string{"b", "c"},
				},
			},
		}
	}
	c := mockCheck()
	c.UnmarshalDataMap()
	c.RunCheck()
	if c.Result.Status != core.Fail {
		t.Errorf("Check should Fail")
	}
	if len(c.Result.Failures) != 1 || c.Result.Failures[0] != "[data] disallowed 'foo': [b, c]" {
		t.Errorf("There should be exactly 1 Failure, got: %#v", c.Result.Failures)
	}

	c = mockCheck()
	c.Values[0].Disallowed = []string{"e"}
	c.UnmarshalDataMap()
	c.RunCheck()
	if c.Result.Status != core.Pass {
		t.Errorf("Check should Pass")
	}
	if len(c.Result.Failures) > 0 {
		t.Errorf("There should be no Failure, got: %#v", c.Result.Failures)
	}

}

func TestYamlCheck(t *testing.T) {
	mockCheck := func() core.YamlCheck {
		return core.YamlCheck{
			YamlBase: core.YamlBase{
				Values: []core.KeyValue{
					{
						Key:   "check.interval_days",
						Value: "7",
					},
				},
			},
			Path: "yaml",
		}
	}

	c := mockCheck()
	c.FetchData()
	if c.Result.Status != core.Fail {
		t.Error("Check with no File or Pattern should Fail")
	}
	if len(c.Result.Failures) != 1 || c.Result.Failures[0] != "no config file name provided" {
		t.Errorf("there should be exactly 1 Failure, got: %#v", c.Result.Failures)
	}

	// Non-existent file.
	c = mockCheck()
	c.File = "non-existent"
	c.FetchData()
	if c.Result.Status != core.Fail {
		t.Error("Check with non-existent file should Fail")
	}
	if len(c.Result.Failures) != 1 || c.Result.Failures[0] != "open testdata/yaml/non-existent.yml: no such file or directory" {
		t.Errorf("there should be exactly 1 Failure, got: %#v", c.Result.Failures)
	}

	// Single file.
	c = mockCheck()
	c.File = "update.settings"
	c.FetchData()
	if len(c.Result.Failures) > 0 {
		t.Errorf("FetchData should succeed, but failed: %+v", c.Result.Failures)
	}
	if !c.HasData(false) {
		t.Errorf("c.DataMap should be filled, but is emptc.")
	}
	c.UnmarshalDataMap()
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

	// Bad File pattern.
	c = mockCheck()
	c.Pattern = "*.bar.yml"
	c.Path = ""
	c.FetchData()
	if c.Result.Status != core.Fail {
		t.Error("Check with bad file pattern should fail")
	}
	if len(c.Result.Failures) != 1 || c.Result.Failures[0] != "error parsing regexp: missing argument to repetition operator: `*`" {
		t.Errorf("there should be exactly 1 Failure, got: %#v", c.Result.Failures)
	}

	// File pattern with no matching files.
	c = mockCheck()
	c.Pattern = "bla.*.yml"
	c.FetchData()
	if c.Result.Status != core.Fail {
		t.Error("Check should Fail")
	}
	if len(c.Result.Failures) != 1 || c.Result.Failures[0] != "no matching config files found" {
		t.Errorf("there should be exactly 1 Failure, got: %#v", c.Result.Failures)
	}

	// Correct file pattern.
	c = mockCheck()
	c.Pattern = ".*.bar.yml"
	c.FetchData()
	if c.Result.Status == core.Fail {
		t.Error("Check should not Fail yet")
	}
	if len(c.Result.Failures) > 0 {
		t.Errorf("there should be no Failure, got: %#v", c.Result.Failures)
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if c.Result.Status != core.Fail {
		t.Error("Check should Fail")
	}
	if len(c.Result.Failures) != 1 || len(c.Result.Passes) != 1 {
		t.Errorf("there should be exactly 1 Failure and 1 Pass, got Failures: %#v, Passes: %#v", c.Result.Failures, c.Result.Passes)
	}
	if c.Result.Failures[0] != "[zoom.bar.yml] 'check.interval_days' equals '5'" || c.Result.Passes[0] != "[foo.bar.yml] 'check.interval_days' equals '7'" {
		t.Errorf("wrong value for Failure or Pass, got Failures: %#v, Passes: %#v", c.Result.Failures, c.Result.Passes)
	}
}
