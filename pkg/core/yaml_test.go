package core_test

import (
	"salsadigitalauorg/shipshape/pkg/core"
	"strings"
	"testing"
)

func TestYamlUnmarshalData(t *testing.T) {
	invalidData := []byte(`
checks:
  drupal-db-config:
	foo
`)
	y := core.YamlCheck{}
	err := y.UnmarshalData(invalidData)
	if err == nil || !strings.Contains(err.Error(), "yaml: line 4: found character that cannot start any token") {
		t.Error("file parsing should fail")
	}

	validData := []byte(`
checks:
  drupal-db-config:
    - name: My db check
      config-name: core.extension
`)
	y = core.YamlCheck{}
	err = y.UnmarshalData(validData)
	if err != nil {
		t.Error("file parsing should succeed")
	}
}

func TestYamlCheckKeyValue(t *testing.T) {
	data := []byte(`
foo:
  bar:
    - baz: zoo

`)
	y := core.YamlCheck{}
	y.UnmarshalData(data)

	// Invalid path.
	_, err := y.CheckKeyValue(core.KeyValue{
		Key:   "&*&^);",
		Value: "foo",
	})
	if err == nil || err.Error() != "child name missing at position 0, following \"\"" {
		t.Error("should fail with error 'child name missing at position 0, following \"\"', got success")
	}

	// Non-existent path.
	kvr, err := y.CheckKeyValue(core.KeyValue{
		Key:   "foo.baz",
		Value: "foo",
	})
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != core.KeyValueNotFound {
		t.Errorf("result should be KeyValueNotFound(-1), got %d", kvr)
	}

	// Wrong value.
	kvr, err = y.CheckKeyValue(core.KeyValue{
		Key:   "foo.bar[0].baz",
		Value: "zoom",
	})
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != core.KeyValueNotEqual {
		t.Errorf("result should be KeyValueNotEqual(0), got %d", kvr)
	}

	// Correct value.
	kvr, err = y.CheckKeyValue(core.KeyValue{
		Key:   "foo.bar[0].baz",
		Value: "zoo",
	})
	if err != nil {
		t.Error("path lookup should succeed")
	}
	if kvr != core.KeyValueEqual {
		t.Errorf("result should be KeyValueEqual(1), got %d", kvr)
	}
}
