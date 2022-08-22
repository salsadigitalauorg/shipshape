package shipshape

import "github.com/salsadigitalauorg/shipshape/pkg/utils"

func MergeKeyValueSlice(kvSlcA *[]KeyValue, kvSlcB []KeyValue) {
	if len(kvSlcB) > 0 {
		*kvSlcA = append(*kvSlcA, kvSlcB...)
	}
}

// CheckAllowDisallowList validates against allow/disallow lists and returns
// true if a disallowed value is present.
func (kv KeyValue) CheckAllowDisallowList(value string) bool {

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
