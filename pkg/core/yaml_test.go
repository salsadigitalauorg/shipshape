package core_test

import (
	"reflect"
	"salsadigitalauorg/shipshape/pkg/core"
	"strings"
	"testing"
)

func TestYamlUnmarshalDataMap(t *testing.T) {
	// Invalid data.
	dataMap := map[string][]byte{
		"data": []byte(`
checks:
  drupal-db-config:
	foo
`),
	}
	y := core.YamlCheck{}
	err := y.UnmarshalDataMap(dataMap)
	if err == nil || !strings.Contains(err.Error(), "yaml: line 4: found character that cannot start any token") {
		t.Error("file parsing should fail")
	}

	// Valid data.
	dataMap["data"] = []byte(`
checks:
  drupal-db-config:
    - name: My db check
      config-name: core.extension
`)
	y = core.YamlCheck{}
	err = y.UnmarshalDataMap(dataMap)
	if err != nil {
		t.Error("file parsing should succeed")
	}
}

func TestYamlCheckKeyValue(t *testing.T) {
	dataMap := map[string][]byte{
		"data": []byte(`
foo:
  bar:
    - baz: zoo

`),
	}

	y := core.YamlCheck{}
	y.UnmarshalDataMap(dataMap)

	// Invalid path.
	_, _, err := y.CheckKeyValue(core.KeyValue{
		Key:   "&*&^);",
		Value: "foo",
	}, "data")
	if err == nil || err.Error() != "child name missing at position 0, following \"\"" {
		t.Error("should fail with error 'child name missing at position 0, following \"\"', got success")
	}

	// Non-existent path.
	kvr, _, err := y.CheckKeyValue(core.KeyValue{
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
	kvr, _, err = y.CheckKeyValue(core.KeyValue{
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
	kvr, _, err = y.CheckKeyValue(core.KeyValue{
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
	dataMap := map[string][]byte{
		"data": []byte(`
foo:
  bar:
    - baz
    - zoo
    - zoom

`),
	}
	y := core.YamlCheck{}
	y.UnmarshalDataMap(dataMap)

	// Disallowed list not provided.
	_, _, err := y.CheckKeyValue(core.KeyValue{
		Key:    "foo.bar",
		IsList: true,
	}, "data")
	if err == nil {
		t.Error("should fail")
	}

	var kvr core.KeyValueResult
	var fails []string
	// Disallowed values in yaml.
	kvr, fails, err = y.CheckKeyValue(core.KeyValue{
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
	kvr, fails, _ = y.CheckKeyValue(core.KeyValue{
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
