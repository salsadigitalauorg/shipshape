package plugin

import "sort"

// GetFactoriesKeys returns a sorted list of keys from a plugin factory registry.
func GetFactoriesKeys[T Plugin](factories interface{}) []string {
	var keys []string

	switch r := factories.(type) {
	case Factories[T]:
		for k := range r {
			keys = append(keys, k)
		}
	case FactoriesNoId[T]:
		for k := range r {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)
	return keys
}

// GetFactory returns a plugin factory from a registry by name.
func GetFactory[T Plugin](factories Factories[T], name string) func(string) T {
	if factory, ok := factories[name]; ok {
		return factory
	}
	var zero func(string) T
	return zero
}

// GetFactoryNoId returns a plugin factory from a registry.
func GetFactoryNoId[T Plugin](factories FactoriesNoId[T], name string) func() T {
	if factory, ok := factories[name]; ok {
		return factory
	}
	var zero func() T
	return zero
}
