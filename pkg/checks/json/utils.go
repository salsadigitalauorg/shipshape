package json

import (
	"github.com/goccy/go-json"
	"github.com/jmespath/go-jmespath"
	"strings"
)

// EvaluateJsonPath evaluates a JSONPath and returns the values.
func EvaluateJsonPath(path string, unmarshalledJsonData any) (foundValues any, err error) {
	var jsonPath *json.Path
	jsonPath, err = json.CreatePath(path)
	if err != nil {
		return nil, err
	}
	err = jsonPath.Get(unmarshalledJsonData, &foundValues)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "failed to get") {
			return nil, err
		}
	}

	return foundValues, nil
}

// EvaluateJMESPath evaluates a JMESPath and returns the values.
func EvaluateJMESPath(path string, unmarshalledJsonData any) (foundValues any, err error) {
	foundValues, err = jmespath.Search(path, unmarshalledJsonData)
	if err != nil {
		return nil, err
	}

	if foundValues != nil {
		if v, ok := foundValues.([]any); ok {
			if len(v) == 0 {
				return nil, err
			}
		}
	}

	return foundValues, nil
}

// LookupJson Lookup from JSON data by either a JSONPath (prefixed with $) or JMESPath.
func LookupJson(path string, unmarshalledJsonData any) (foundValues any, err error, pathType string) {
	if strings.HasPrefix(path, "$") {
		foundValues, err = EvaluateJsonPath(path, unmarshalledJsonData)
		pathType = "JSONPath"
	} else {
		foundValues, err = EvaluateJMESPath(path, unmarshalledJsonData)
		pathType = "JMESPath"
	}

	return foundValues, err, pathType
}
