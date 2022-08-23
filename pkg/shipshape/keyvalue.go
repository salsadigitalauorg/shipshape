package shipshape

import "github.com/salsadigitalauorg/shipshape/pkg/utils"

func MergeKeyValueSlice(kvSlcA *[]KeyValue, kvSlcB []KeyValue) {
	if len(kvSlcB) == 0 {
		return
	}
	newKvSlc := *kvSlcA
	for _, kvB := range kvSlcB {
		kvA, indA := getKeyValueFromSlice(kvSlcA, kvB.Key)
		if kvA == nil {
			*kvSlcA = append(*kvSlcA, kvB)
			continue
		}
		newKvSlc[indA] = kvB
	}
	kvSlcA = &newKvSlc
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
