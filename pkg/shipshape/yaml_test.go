package shipshape_test

import (
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func TestYamlUnmarshalDataMap(t *testing.T) {
	// Invalid data.
	c := shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
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
	if _, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error("invalid yaml data should Fail")
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"yaml: line 4: found character that cannot start any token"}); !ok {
		t.Error(msg)
	}

	// Valid data.
	c = shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
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
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	// Invalid yaml kec.
	c = shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    baz&*zoom: zap
`),
			},
		},
		Values: []shipshape.KeyValue{
			{Key: "baz&*zoom", Value: "zap"},
		},
	}
	c.RunCheck()
	if _, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error("invalid yaml key should Fail")
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"invalid character '&' at position 3, following \"baz\""}); !ok {
		t.Error(msg)
	}
}

func TestYamlCheckKeyValue(t *testing.T) {
	c := shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
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
	_, _, err := c.CheckKeyValue(shipshape.KeyValue{
		Key:   "&*&^);",
		Value: "foo",
	}, "data")
	if err == nil || err.Error() != "child name missing at position 0, following \"\"" {
		t.Error("should fail with error 'child name missing at position 0, following \"\"', got success")
	}

	// Non-existent path.
	kvr, _, err := c.CheckKeyValue(shipshape.KeyValue{
		Key:   "foo.baz",
		Value: "foo",
	}, "data")
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != shipshape.KeyValueNotFound {
		t.Errorf("result should be KeyValueNotFound(-1), got %d", kvr)
	}

	// Wrong value.
	kvr, _, err = c.CheckKeyValue(shipshape.KeyValue{
		Key:   "foo.bar[0].baz",
		Value: "zoom",
	}, "data")
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != shipshape.KeyValueNotEqual {
		t.Errorf("result should be KeyValueNotEqual(0), got %d", kvr)
	}

	// Correct value.
	kvr, _, err = c.CheckKeyValue(shipshape.KeyValue{
		Key:   "foo.bar[0].baz",
		Value: "zoo",
	}, "data")
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != shipshape.KeyValueEqual {
		t.Errorf("result should be KeyValueEqual(1), got %d", kvr)
	}

	// Optional value not present.
	kvr, _, err = c.CheckKeyValue(shipshape.KeyValue{
		Key:   "foo.bar[0].bazzle",
		Value: "zoom",
		Optional: true,
	}, "data")
	if err != nil {
		t.Error("missing optional value should not fail")
	}
	if kvr != shipshape.KeyValueEqual {
		t.Errorf("result should be KeyValueEqual(1), got %d", kvr)
	}
}

func TestYamlCheckKeyValueList(t *testing.T) {
	c := shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
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
	_, _, err := c.CheckKeyValue(shipshape.KeyValue{
		Key:    "foo.bar",
		IsList: true,
	}, "data")
	if err == nil {
		t.Error("should fail")
	}

	var kvr shipshape.KeyValueResult
	var fails []string
	// Disallowed values in yaml.
	kvr, fails, err = c.CheckKeyValue(shipshape.KeyValue{
		Key:        "foo.bar",
		IsList:     true,
		Disallowed: []string{"baz", "zoo"},
	}, "data")
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != shipshape.KeyValueDisallowedFound {
		t.Errorf("result should be KeyValueDisallowedFound(-2), got %d", kvr)
	}
	expectedFails := []string{"baz", "zoo"}
	if len(fails) != 2 || !reflect.DeepEqual(fails, expectedFails) {
		t.Errorf("There should be exactly 2 Failures, with values %+v, got %+v", expectedFails, fails)
	}

	// No disallowed values in yaml.
	kvr, fails, _ = c.CheckKeyValue(shipshape.KeyValue{
		Key:        "foo.bar",
		IsList:     true,
		Disallowed: []string{"this should", "be a success"},
	}, "data")
	if kvr != shipshape.KeyValueEqual {
		t.Errorf("result should be KeyValueEqual(1), got %d", kvr)
	}
	if len(fails) > 0 {
		t.Errorf("There should be no Failures, got %+v", fails)
	}

	// Allowed values in yaml all match.
	kvr, fails, _ = c.CheckKeyValue(shipshape.KeyValue{
		Key:        "foo.bar",
		IsList:     true,
		Allowed: []string{"baz", "zoo", "zoom"},
	}, "data")
	if kvr != shipshape.KeyValueEqual {
		t.Errorf("result should be KeyValueEqual(1), got %d", kvr)
	}
	if len(fails) > 0 {
		t.Errorf("There should be no Failures, got %+v", fails)
	}

	// Value not in Allowed list.
	kvr, fails, _ = c.CheckKeyValue(shipshape.KeyValue{
		Key:        "foo.bar",
		IsList:     true,
		Allowed: []string{"baz", "zoo"},
	}, "data")
	if kvr != shipshape.KeyValueDisallowedFound {
		t.Errorf("result should be KeyValueDisallowedFound(-2), got %d", kvr)
	}
	expectedAllowedFails := []string{"zoom"}
	if len(fails) != 1 || !reflect.DeepEqual(fails, expectedAllowedFails) {
		t.Errorf("There should be exactly 1 Failure, with values %+v, got %+v", expectedAllowedFails, fails)
	}

}

func TestYamlBase(t *testing.T) {
	c := shipshape.YamlBase{}
	c.HasData(true)
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"no data available"}); !ok {
		t.Error(msg)
	}

	mockCheck := func() shipshape.YamlBase {
		return shipshape.YamlBase{
			CheckBase: shipshape.CheckBase{
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
			Values: []shipshape.KeyValue{
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
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"[data] 'check.interval_days' equals '7'"}); !ok {
		t.Error(msg)
	}

	// Wrong key, correct value.
	c = mockCheck()
	c.Values = []shipshape.KeyValue{
		{
			Key:   "check.interval",
			Value: "7",
		},
	}
	c.RunCheck()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"[data] 'check.interval' not found"}); !ok {
		t.Error(msg)
	}

	// Correct key, wrong value.
	c = mockCheck()
	c.Values = []shipshape.KeyValue{
		{
			Key:   "check.interval_days",
			Value: "8",
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"[data] 'check.interval_days' equals '7'"}); !ok {
		t.Error(msg)
	}

	// Multiple config values - all correct.
	c = mockCheck()
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
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"[data] 'check.interval_days' equals '7'",
		"[data] 'notification.emails[0]' equals 'admin@example.com'",
	}); !ok {
		t.Error(msg)
	}

	// Wildcard key.
	c = mockCheck()
	c.DataMap = map[string][]byte{
		"data": []byte(`
abcd:
  some:
    - thing 1
    - thing 2
    - thing 3
efgh:
  some:
    - thing 1
    - thing 2
    - thing 3
`),
	}
	c.Values = []shipshape.KeyValue{
		{
			Key:        "*.some",
			IsList:     true,
			Disallowed: []string{"thing 2", "thing 4"},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"[data] disallowed *.some: [thing 2]"}); !ok {
		t.Error(msg)
	}
}

func TestYamlBaseListValues(t *testing.T) {
	mockCheck := func() shipshape.YamlBase {
		return shipshape.YamlBase{
			CheckBase: shipshape.CheckBase{
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
			Values: []shipshape.KeyValue{
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
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"[data] disallowed foo: [b, c]"}); !ok {
		t.Error(msg)
	}

	c = mockCheck()
	c.Values[0].Disallowed = []string{"e"}
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"[data] no disallowed 'foo'"}); !ok {
		t.Error(msg)
	}

}

func TestYamlCheck(t *testing.T) {
	mockCheck := func() shipshape.YamlCheck {
		return shipshape.YamlCheck{
			YamlBase: shipshape.YamlBase{
				Values: []shipshape.KeyValue{
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
	if _, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error("Check with no File or Pattern should Fail")
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"no file provided"}); !ok {
		t.Error(msg)
	}

	// Non-existent file.
	c = mockCheck()
	c.Init("testdata", shipshape.Yaml)
	c.File = "non-existent.yml"
	c.FetchData()
	if _, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error("Check with non-existent file should Fail")
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"open testdata/yaml/non-existent.yml: no such file or directory"}); !ok {
		t.Error(msg)
	}

	// Non-existent file with ignore missing.
	c = mockCheck()
	c.File = "non-existent.yml"
	c.IgnoreMissing = true
	c.FetchData()
	if _, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error("Check with non-existent file when ignoring missing should Pass already.")
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"File testdata/yaml/non-existent.yml does not exist"}); !ok {
		t.Error(msg)
	}

	// Single file.
	c = mockCheck()
	c.File = "update.settings.yml"
	c.FetchData()
	if msg, ok := internal.EnsureNoFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if !c.HasData(false) {
		t.Errorf("c.DataMap should be filled, but is empty.")
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"[yaml/update.settings.yml] 'check.interval_days' equals '7'"}); !ok {
		t.Error(msg)
	}

	// Bad File pattern.
	c = mockCheck()
	c.Pattern = "*.bar.yml"
	c.Path = ""
	c.FetchData()
	if _, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error("Check with bad file pattern should fail")
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"error parsing regexp: missing argument to repetition operator: `*`"}); !ok {
		t.Error(msg)
	}

	// File pattern with no matching files.
	c = mockCheck()
	c.Pattern = "bla.*.yml"
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"no matching config files found"}); !ok {
		t.Error(msg)
	}

	// File pattern with no matching files, ignoring missing.
	c = mockCheck()
	c.Pattern = "bla.*.yml"
	c.IgnoreMissing = true
	c.FetchData()
	if _, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error("Check with non-existent file pattern when ignoring missing should Pass already.")
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"no matching config files found"}); !ok {
		t.Error(msg)
	}

	// Correct file pattern.
	c = mockCheck()
	c.Pattern = ".*.bar.yml"
	c.FetchData()
	if c.Result.Status == shipshape.Fail {
		t.Error("Check should not Fail yet")
	}
	if len(c.Result.Failures) > 0 {
		t.Errorf("there should be no Failure, got: %#v", c.Result.Failures)
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"[yaml/foo.bar.yml] 'check.interval_days' equals '7'"}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"[yaml/zoom.bar.yml] 'check.interval_days' equals '5'"}); !ok {
		t.Error(msg)
	}
}

func TestYamlLintCheck(t *testing.T) {
	mockCheck := func(file string, files []string, ignoreMissing bool) shipshape.YamlLintCheck {
		return shipshape.YamlLintCheck{
			YamlCheck: shipshape.YamlCheck{
				YamlBase: shipshape.YamlBase{
					CheckBase: shipshape.CheckBase{
						Name:    "Test yaml lint",
						DataMap: map[string][]byte{},
					},
				},
				File:          file,
				Files:         files,
				IgnoreMissing: ignoreMissing,
			},
		}
	}

	c := mockCheck("", []string{}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"no file provided"}); !ok {
		t.Error(msg)
	}

	c = mockCheck("non-existent-file.yml", []string{}, true)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureNoFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"File testdata/non-existent-file.yml does not exist"}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	c = mockCheck("", []string{"non-existent-file.yml", "yaml-invalid.yml"}, true)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureNoFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"File testdata/non-existent-file.yml does not exist",
		"File testdata/yaml-invalid.yml does not exist",
	}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	c = mockCheck("non-existent-file.yml", []string{}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"open testdata/non-existent-file.yml: no such file or directory"}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	c = mockCheck("", []string{"non-existent-file.yml", "yamllint-invalid.yml"}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"open testdata/non-existent-file.yml: no such file or directory"}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	c = mockCheck("", []string{}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.DataMap["yaml-invalid.yml"] = []byte(`
this: is invalid
this: yaml
`)
	c.UnmarshalDataMap()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"[yaml-invalid.yml] line 3: mapping key \"this\" already defined at line 2"}); !ok {
		t.Error(msg)
	}

	c = mockCheck("", []string{}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.DataMap["yaml-valid.yml"] = []byte(`
this: is
valid: yaml
`)
	c.UnmarshalDataMap()
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"yaml-valid.yml has valid yaml."}); !ok {
		t.Error(msg)
	}
}
