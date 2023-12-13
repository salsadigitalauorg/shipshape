package json_test

import (
	"github.com/goccy/go-json"
	. "github.com/salsadigitalauorg/shipshape/pkg/checks/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

// GetTestJsonBlob return sample JSON for testing.
func GetTestJsonBlob() []byte {
	return []byte(`{
    "require": {
        "awesome/library": "dev-my-fork-awesome-feature"
    },
    "repositories" : [
        {
            "type" : "vcs",
            "url" : "https://github.com/yourgithubuser/library",
            "no-api": true
        },
		{
            "type" : "vcs",
            "url" : "https://github.com/yourgithubuser/library1",
			"custom": "Custom data"
        },
		{
            "type" : "vcs",
            "url" : "https://github.com/yourgithubuser/library2",
			"no-api": false
        }
    ]
}`)
}

// UnmarshalJson parses a JSON string.
func UnmarshalJson(jsonBlob []byte) (any, error) {
	var unmarshalledJsonData any
	err := json.Unmarshal(jsonBlob, &unmarshalledJsonData)
	return unmarshalledJsonData, err
}

// TestEvaluateJsonPath tests EvaluateJsonPath function.
func TestEvaluateJsonPath(t *testing.T) {
	assertions := assert.New(t)
	jsonBlob := GetTestJsonBlob()
	unmarshalledJsonData, jsonError := UnmarshalJson(jsonBlob)
	assertions.Nil(jsonError)

	foundNodes, err := EvaluateJsonPath("$.require['awesome/library']", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Equal(foundNodes.(string), "dev-my-fork-awesome-feature")

	foundNodes, err = EvaluateJsonPath("$.require['awesome/invalid-library']", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.Nil(foundNodes)

	foundNodes, err = EvaluateJsonPath("$.repositories..no-api", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Len(foundNodes, 2)
	assertions.Equal(foundNodes.([]any)[0], true)
	assertions.Equal(foundNodes.([]any)[1], false)
	assertions.EqualValues(foundNodes.([]any), []any{true, false})

	foundNodes, err = EvaluateJsonPath("$.repositories..type", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Len(foundNodes, 3)
	assertions.EqualValues(foundNodes.([]any), []any{"vcs", "vcs", "vcs"})

	foundNodes, err = EvaluateJsonPath("$.repositories..custom", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Len(foundNodes, 1)
	assertions.EqualValues(foundNodes.([]any), []any{"Custom data"})

	foundNodes, err = EvaluateJsonPath("$.repositories..invalid", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.Nil(foundNodes)

	foundNodes, err = EvaluateJsonPath("repositories.invalidPath", unmarshalledJsonData)
	assertions.NotNil(err)
	assertions.Nil(foundNodes)
}

// TestEvaluateJMESPath tests EvaluateJMESPath function.
func TestEvaluateJMESPath(t *testing.T) {
	assertions := assert.New(t)
	jsonBlob := GetTestJsonBlob()
	unmarshalledJsonData, jsonError := UnmarshalJson(jsonBlob)
	assertions.Nil(jsonError)

	foundNodes, err := EvaluateJMESPath(`require."awesome/library"`, unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Equal(foundNodes.(string), "dev-my-fork-awesome-feature")

	foundNodes, err = EvaluateJMESPath(`repositories[*]."no-api"`, unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Len(foundNodes, 2)
	assertions.Equal(foundNodes.([]any)[0], true)
	assertions.Equal(foundNodes.([]any)[1], false)
	assertions.EqualValues(foundNodes.([]any), []any{true, false})

	foundNodes, err = EvaluateJMESPath("repositories[*].type", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Len(foundNodes, 3)
	assertions.EqualValues(foundNodes.([]any), []any{"vcs", "vcs", "vcs"})

	foundNodes, err = EvaluateJMESPath("repositories[*].custom", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Len(foundNodes, 1)
	assertions.EqualValues(foundNodes.([]any), []any{"Custom data"})

	foundNodes, err = EvaluateJMESPath("repositories[*].invalid", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.Nil(foundNodes)

	foundNodes, err = EvaluateJMESPath("repositories[invalid\\path]", unmarshalledJsonData)
	assertions.NotNil(err)
	assertions.Nil(foundNodes)
}

// TestLookupJson tests LookupJson function.
func TestLookupJson(t *testing.T) {
	assertions := assert.New(t)
	jsonBlob := GetTestJsonBlob()
	unmarshalledJsonData, jsonError := UnmarshalJson(jsonBlob)
	assertions.Nil(jsonError)

	foundNodes, err, pathType := LookupJson(`require."awesome/library"`, unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Equal(foundNodes.(string), "dev-my-fork-awesome-feature")
	assertions.Equal(pathType, "JMESPath")

	foundNodes, err, pathType = LookupJson(`$.require['awesome/library']`, unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Equal(foundNodes.(string), "dev-my-fork-awesome-feature")
	assertions.Equal(pathType, "JSONPath")

	foundNodes, err, pathType = LookupJson(`repositories[*]."no-api"`, unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Len(foundNodes, 2)
	assertions.Equal(foundNodes.([]any)[0], true)
	assertions.Equal(foundNodes.([]any)[1], false)
	assertions.EqualValues(foundNodes.([]any), []any{true, false})
	assertions.Equal(pathType, "JMESPath")

	foundNodes, err, pathType = LookupJson("$.repositories..no-api", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.NotNil(foundNodes)
	assertions.Len(foundNodes, 2)
	assertions.Equal(foundNodes.([]any)[0], true)
	assertions.Equal(foundNodes.([]any)[1], false)
	assertions.EqualValues(foundNodes.([]any), []any{true, false})
	assertions.Equal(pathType, "JSONPath")

	foundNodes, err, pathType = LookupJson(`require."awesome/invalid-library"`, unmarshalledJsonData)
	assertions.Nil(err)
	assertions.Nil(foundNodes)
	assertions.Equal(pathType, "JMESPath")

	foundNodes, err, pathType = LookupJson("$.require['awesome/invalid-library']", unmarshalledJsonData)
	assertions.Nil(err)
	assertions.Nil(foundNodes)
	assertions.Equal(pathType, "JSONPath")

	foundNodes, err, pathType = LookupJson("$.require[awesome/invalid-library]", unmarshalledJsonData)
	assertions.NotNil(err)
	assertions.Nil(foundNodes)
	assertions.Equal(pathType, "JSONPath")

	foundNodes, err, pathType = LookupJson(`require.awesome/invalid-library`, unmarshalledJsonData)
	assertions.NotNil(err)
	assertions.Nil(foundNodes)
	assertions.Equal(pathType, "JMESPath")
}
