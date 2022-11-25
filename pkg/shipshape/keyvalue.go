package shipshape

import "github.com/salsadigitalauorg/shipshape/pkg/utils"

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
