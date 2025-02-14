package plugin

import "sort"

// GetRegistryKeys returns a sorted list of keys from a plugin registry.
func GetRegistryKeys[T Plugin](registry interface{}) []string {
	var keys []string

	switch r := registry.(type) {
	case Registry[T]:
		for k := range r {
			keys = append(keys, k)
		}
	case RegistryNoId[T]:
		for k := range r {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)
	return keys
}

// GetInstance returns a plugin instance from a registry by name.
func GetInstance[T Plugin](registry Registry[T], name string) T {
	if plugin, ok := registry[name]; ok {
		return plugin(name)
	}
	var zero T
	return zero
}

// GetInstanceNoId returns a plugin instance from a registry.
func GetInstanceNoId[T Plugin](registry RegistryNoId[T], name string) T {
	if plugin, ok := registry[name]; ok {
		return plugin()
	}
	var zero T
	return zero
}
