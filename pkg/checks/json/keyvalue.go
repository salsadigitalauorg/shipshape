package json

import (
	"fmt"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
	"reflect"
	"strings"
)

type KeyValue struct {
	yaml.KeyValue    `yaml:",inline"`
	DisallowedValues []any `yaml:"disallowed-values"`
	AllowedValues    []any `yaml:"allowed-values"`
}

// IsDisallowed validates against the allowed/disallowed value lists and returns
// true if a disallowed value is present.
func (kv KeyValue) IsDisallowed(value any) bool {
	// Ignore blank and null values.
	if kv.IsEmpty(value) {
		return false
	}

	// Check disallowed list.
	if len(kv.DisallowedValues) > 0 && utils.SliceContains(kv.DisallowedValues, value) {
		return true
	}

	// Check allowed list.
	if len(kv.AllowedValues) > 0 && !utils.SliceContains(kv.AllowedValues, value) {
		return true
	}

	return false
}

// IsEmpty determines if a value if empty.
func (kv KeyValue) IsEmpty(value any) bool {
	if value == nil {
		return true
	}

	rt := reflect.ValueOf(value)
	switch rt.Kind() {
	case reflect.String:
		return len(rt.String()) == 0
	case reflect.Array, reflect.Slice:
		return rt.Len() == 0
	case reflect.Map:
		return len(rt.MapKeys()) == 0
	default:
		return false
	}
}

// Equals simply returns whether the given value matches the KeyValue.
func (kv KeyValue) Equals(value any) bool {
	if kv.Truthy {
		return kv.EqualsTruthy(fmt.Sprint(value))
	}

	if v, ok := value.(string); ok {
		return strings.EqualFold(kv.Value, v)
	}
	return kv.Value == fmt.Sprint(value)
}
