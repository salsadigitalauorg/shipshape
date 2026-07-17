package config

import (
	log "github.com/sirupsen/logrus"
)

// deepMerge merges src into dst in place and returns the result.
//
// Merge semantics:
//   - When a key holds a map on both sides, the maps are merged recursively.
//   - When a key is absent in dst, src's value is set.
//   - For any other case (scalar or slice on either side), src's value
//     replaces dst's value. Slices are never appended or unioned.
//
// Whenever a leaf or slice value in dst is overwritten, an override is logged
// at Warn with the dotted key path and the originating source, so that a
// later-merged file silently changing a check is never invisible.
func deepMerge(dst, src map[string]interface{}, path, source string) map[string]interface{} {
	if dst == nil {
		dst = map[string]interface{}{}
	}

	for k, srcVal := range src {
		keyPath := k
		if path != "" {
			keyPath = path + "." + k
		}

		dstVal, exists := dst[k]
		if !exists {
			dst[k] = srcVal
			continue
		}

		srcMap, srcIsMap := srcVal.(map[string]interface{})
		dstMap, dstIsMap := dstVal.(map[string]interface{})
		if srcIsMap && dstIsMap {
			dst[k] = deepMerge(dstMap, srcMap, keyPath, source)
			continue
		}

		log.WithField("key", keyPath).
			WithField("source", source).
			Warn("config value overridden during merge")
		dst[k] = srcVal
	}

	return dst
}
