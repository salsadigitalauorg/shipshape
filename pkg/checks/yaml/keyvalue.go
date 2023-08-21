package yaml

import (
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// KeyValue represents a check to be made against Yaml data.
// It can be a simple Key=Value check, or match against a list of Disallowed or
// Allowed values. If the source is a list then IsList must be true.
// If Optional is set then the validation will not fail if the key is not present.
type KeyValue struct {
	Key        string   `yaml:"key"`
	Value      string   `yaml:"value"`
	Truthy     bool     `yaml:"truthy"`
	IsList     bool     `yaml:"is-list"`
	Optional   bool     `yaml:"optional"`
	Disallowed []string `yaml:"disallowed"`
	Allowed    []string `yaml:"allowed"`
}

// KeyValueResult represents the different outcomes of the KeyValue check.
type KeyValueResult int8

const (
	KeyValueError           KeyValueResult = -2
	KeyValueNotFound        KeyValueResult = -1
	KeyValueNotEqual        KeyValueResult = 0
	KeyValueEqual           KeyValueResult = 1
	KeyValueDisallowedFound KeyValueResult = 2
)

var truthyValues = []string{"1", "true"}
var falsyValues = []string{"0", "false", "null"}

// MergeIntSlice replaces the values of a KeyValue slice with those of another.
func MergeKeyValueSlice(slcA *[]KeyValue, slcB []KeyValue) {
	if len(slcB) == 0 {
		return
	}
	// Create new slice with unique values.
	newSlc := []KeyValue{}
	for _, elB := range slcB {
		if kvB, _ := getKeyValueFromSlice(&newSlc, elB.Key); kvB == nil {
			newSlc = append(newSlc, elB)
		}
	}
	*slcA = newSlc
	if len(slcB) == 0 {
		return
	}
}

func getKeyValueFromSlice(kvSlc *[]KeyValue, key string) (*KeyValue, int) {
	for i, kv := range *kvSlc {
		if kv.Key == key {
			return &kv, i
		}
	}
	return nil, -1
}

// Equals simply returns whether the given value matches the keyvalue.
func (kv KeyValue) Equals(value string) bool {
	if kv.Truthy {
		return kv.EqualsTruthy(value)
	}
	return kv.Value == value
}

func (kv KeyValue) EqualsTruthy(value string) bool {
	// Check if true.
	if utils.StringSliceContains(truthyValues, kv.Value) {
		return utils.StringSliceContains(truthyValues, value)
	}
	// Check if false.
	if utils.StringSliceContains(falsyValues, kv.Value) {
		return utils.StringSliceContains(falsyValues, value)
	}
	return false
}

// IsDisallowed validates against the allow/disallow lists and returns
// true if a disallowed value is present.
func (kv KeyValue) IsDisallowed(value string) bool {

	// Ignore blank and null values.
	if len(value) == 0 {
		return false
	}

	// Check disallowed list.
	if len(kv.Disallowed) > 0 && utils.StringSliceContains(kv.Disallowed, value) {
		return true
	}

	// Check allowed list.
	if len(kv.Allowed) > 0 && !utils.StringSliceContains(kv.Allowed, value) {
		return true
	}

	return false
}
