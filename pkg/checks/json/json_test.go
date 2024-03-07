package json_test

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/checks/json"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

func TestJsonCheckUnmarshalDataMap(t *testing.T) {
	assertions := assert.New(t)

	// Invalid data.
	c := JsonCheck{
		YamlCheck: yaml.YamlCheck{
			YamlBase: yaml.YamlBase{
				CheckBase: config.CheckBase{
					DataMap: map[string][]byte{
						"data": []byte(`{
	"repositories": {
		package1: "package1",
	}
}`),
					},
				},
			},
		},
	}

	c.UnmarshalDataMap()
	assertions.EqualValues(0, len(c.Result.Passes))
	assertions.ElementsMatch(
		[]breach.Breach{
			&breach.ValueBreach{
				BreachType: breach.BreachTypeValue,
				ValueLabel: "JSON error",
				Value:      "invalid character 'p' looking for beginning of value",
			},
		},
		c.Result.Breaches,
	)

	// Valid data.
	c = JsonCheck{
		YamlCheck: yaml.YamlCheck{
			YamlBase: yaml.YamlBase{
				CheckBase: config.CheckBase{
					DataMap: map[string][]byte{
						"data": []byte(`{
	"repositories": {
		"package1": "package1",
		"package2": "package2"
	}
}`),
					},
				},
			},
		},
	}

	c.UnmarshalDataMap()
	assertions.EqualValues(0, len(c.Result.Breaches))
}

func TestJsonCheckKeyValue(t *testing.T) {
	var singleValueNode any
	json.Unmarshal([]byte(`{
	"license": "MIT",
	"repositories": {
		"package1": {
			"type": "vcs",
			"url": "https://github.com/yourgithubuser/package1"
		}
	}
}`), &singleValueNode)

	var multiValueNode any
	json.Unmarshal([]byte(`{
	"license": "BSD",
	"repositories": {
		"package1": {
			"type": "vcs",
			"url": "https://github.com/yourgithubuser/package1",
			"no-api": true
		},
		"package2": {
			"type": "library",
			"url": "https://github.com/yourgithubuser/package2",
			"no-api": false
		}
	}
}`), &multiValueNode)

	tests := []struct {
		name           string
		node           any
		keyValue       KeyValue
		expectedResult yaml.KeyValueResult
		expectedValues []string
		expectedError  string
	}{
		{
			name: "invalid JSONPath",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:   "$.**);",
					Value: "foo",
				},
			},
			expectedResult: yaml.KeyValueError,
			expectedValues: nil,
			expectedError:  "json: invalid path format: found invalid path character * after dot",
		},
		{
			name: "invalid JMESPath",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:   "**);",
					Value: "foo",
				},
			},
			expectedResult: yaml.KeyValueError,
			expectedValues: nil,
			expectedError:  "SyntaxError: Unknown char: ';'",
		},
		{
			name: "non-existent JSONPath",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:   "$.authors",
					Value: "foo",
				},
			},
			expectedResult: yaml.KeyValueNotFound,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "non-existent JMESPath",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:   "authors",
					Value: "foo",
				},
			},
			expectedResult: yaml.KeyValueNotFound,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "JSONPath optional value not present",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:      "$.authors",
					Value:    "foo",
					Optional: true,
				},
			},
			expectedResult: yaml.KeyValueEqual,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "JMESPath optional value not present",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:      "authors",
					Value:    "foo",
					Optional: true,
				},
			},
			expectedResult: yaml.KeyValueEqual,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "JSONPath wrong value",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:   "$.license",
					Value: "BSD",
				},
			},
			expectedResult: yaml.KeyValueNotEqual,
			expectedValues: []string{"MIT"},
			expectedError:  "",
		},
		{
			name: "JMESPath wrong value",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:   "license",
					Value: "BSD",
				},
			},
			expectedResult: yaml.KeyValueNotEqual,
			expectedValues: []string{"MIT"},
			expectedError:  "",
		},
		{
			name: "JSONPath correct value",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:   "$.license",
					Value: "MIT",
				},
			},
			expectedResult: yaml.KeyValueEqual,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "JMESPath correct value",
			node: singleValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:   "license",
					Value: "MIT",
				},
			},
			expectedResult: yaml.KeyValueEqual,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "JSONPath multivalue - is-list not set",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key: "$.repositories..type",
				},
			},
			expectedResult: yaml.KeyValueError,
			expectedValues: nil,
			expectedError:  "A list of values was found but is-list is not set",
		},
		{
			name: "JMESPath multivalue - is-list not set",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key: "repositories.*.type",
				},
			},
			expectedResult: yaml.KeyValueError,
			expectedValues: nil,
			expectedError:  "A list of values was found but is-list is not set",
		},
		{
			name: "JSONPath multivalue - disallowed list not provided",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "$.repositories..type",
					IsList: true,
				},
			},
			expectedResult: yaml.KeyValueError,
			expectedValues: nil,
			expectedError:  "list of allowed or disallowed values not provided",
		},
		{
			name: "JMESPath multivalue - disallowed list not provided",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "repositories.*.type",
					IsList: true,
				},
			},
			expectedResult: yaml.KeyValueError,
			expectedValues: nil,
			expectedError:  "list of allowed or disallowed values not provided",
		},
		{
			name: "JSONPath multivalue - disallowed value found",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "$.repositories..type",
					IsList: true,
				},
				DisallowedValues: []any{"vcs", "library", "project"},
			},
			expectedResult: yaml.KeyValueDisallowedFound,
			expectedValues: []string{"vcs", "library"},
			expectedError:  "",
		},
		{
			name: "JMESPath multivalue - disallowed value found",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "repositories.*.type",
					IsList: true,
				},
				DisallowedValues: []any{"vcs", "library", "project"},
			},
			expectedResult: yaml.KeyValueDisallowedFound,
			expectedValues: []string{"vcs", "library"},
			expectedError:  "",
		},
		{
			name: "JSONPath multivalue - no disallowed value found",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "$.repositories..type",
					IsList: true,
				},
				DisallowedValues: []any{"composer-plugin", "project"},
			},
			expectedResult: yaml.KeyValueEqual,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "JMESPath multivalue - no disallowed value found",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "repositories.*.type",
					IsList: true,
				},
				DisallowedValues: []any{"composer-plugin", "project"},
			},
			expectedResult: yaml.KeyValueEqual,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "JSONPath multivalue - allowed values matched",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "$.repositories..type",
					IsList: true,
				},
				AllowedValues: []any{"vcs", "library"},
			},
			expectedResult: yaml.KeyValueEqual,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "JMESPath multivalue - allowed values matched",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "repositories.*.type",
					IsList: true,
				},
				AllowedValues: []any{"vcs", "library"},
			},
			expectedResult: yaml.KeyValueEqual,
			expectedValues: nil,
			expectedError:  "",
		},
		{
			name: "JSONPath multivalue - value not in allowed list",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "$.repositories..type",
					IsList: true,
				},
				AllowedValues: []any{"composer-plugin", "project"},
			},
			expectedResult: yaml.KeyValueDisallowedFound,
			expectedValues: []string{"vcs", "library"},
			expectedError:  "",
		},
		{
			name: "JMESPath multivalue - value not in allowed list",
			node: multiValueNode,
			keyValue: KeyValue{
				KeyValue: yaml.KeyValue{
					Key:    "repositories.*.type",
					IsList: true,
				},
				AllowedValues: []any{"composer-plugin", "project"},
			},
			expectedResult: yaml.KeyValueDisallowedFound,
			expectedValues: []string{"vcs", "library"},
			expectedError:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)

			kvr, values, err := CheckKeyValue(test.node, test.keyValue)
			assertions.Equal(test.expectedResult, kvr, "expected result to match")
			assertions.ElementsMatch(test.expectedValues, values, "expected values to match")
			if err != nil {
				assertions.Equal(err.Error(), test.expectedError, "expected error to match")
			} else if test.expectedError != "" {
				assertions.Fail(test.expectedError, "expected error to match")
			}
		})
	}

}
