package yaml_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestYamlBaseMerge(t *testing.T) {
	assert := assert.New(t)

	c := YamlBase{
		Values: []KeyValue{{Key: "foo", Value: "bar"}},
	}
	err := c.Merge(&YamlBase{
		Values: []KeyValue{{Key: "baz", Value: "zoom"}},
	})
	assert.Equal(nil, err)
	assert.EqualValues(
		[]KeyValue{{Key: "baz", Value: "zoom"}},
		c.Values,
	)
	err = c.Merge(&YamlBase{
		Values: []KeyValue{{Key: "baz", Value: "zap"}},
	})
	assert.Equal(nil, err)
	assert.EqualValues(
		[]KeyValue{{Key: "baz", Value: "zap"}},
		c.Values,
	)
}

func TestYamlUnmarshalDataMap(t *testing.T) {
	assert := assert.New(t)

	// Invalid data.
	c := YamlBase{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
	baz
`),
			},
		},
	}
	c.UnmarshalDataMap()
	assert.EqualValues(0, len(c.Result.Passes))
	assert.ElementsMatch([]breach.Breach{&breach.ValueBreach{
		BreachType: breach.BreachTypeValue,
		Value:      "yaml: line 4: found character that cannot start any token"}},
		c.Result.Breaches)

	// Valid data.
	c = YamlBase{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    - name: baz
      value: zoom
`),
			},
		},
	}
	c.UnmarshalDataMap()
	assert.EqualValues(0, len(c.Result.Breaches))

	// Invalid yaml kec.
	c = YamlBase{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    baz&*zoom: zap
`),
			},
		},
		Values: []KeyValue{
			{Key: "baz&*zoom", Value: "zap"},
		},
	}
	c.RunCheck()
	c.Result.DetermineResultStatus(false)

	assert.Equal(result.Fail, c.Result.Status)
	assert.ElementsMatch([]breach.Breach{&breach.ValueBreach{
		BreachType: breach.BreachTypeValue,
		Value:      "invalid character '&' at position 3, following \"baz\""}},
		c.Result.Breaches)
}

func TestYamlCheckKeyValue(t *testing.T) {
	singleValueNode := yamlv3.Node{}
	yamlv3.Unmarshal([]byte(`
foo:
  bar:
    - baz: zoo
    - zap: Bam
`), &singleValueNode)

	multiValueNode := yamlv3.Node{}
	yamlv3.Unmarshal([]byte(`
foo:
  bar:
    - baz
    - zoo
    - zoom
`), &multiValueNode)

	tests := []struct {
		name           string
		node           yamlv3.Node
		keyValue       KeyValue
		expectedResult KeyValueResult
		expectedValues []string
		expectedError  error
	}{
		{
			name: "invalid path",
			node: singleValueNode,
			keyValue: KeyValue{
				Key:   "&*&^);",
				Value: "foo",
			},
			expectedResult: KeyValueError,
			expectedValues: nil,
			expectedError:  errors.New("child name missing at position 0, following \"\""),
		},
		{
			name: "non-existent path",
			node: singleValueNode,
			keyValue: KeyValue{
				Key:   "foo.baz",
				Value: "foo",
			},
			expectedResult: KeyValueNotFound,
			expectedValues: nil,
			expectedError:  nil,
		},
		{
			name: "wrong value",
			node: singleValueNode,
			keyValue: KeyValue{
				Key:   "foo.bar[0].baz",
				Value: "zoom",
			},
			expectedResult: KeyValueNotEqual,
			expectedValues: []string{"zoo"},
			expectedError:  nil,
		},
		{
			name: "correct value",
			node: singleValueNode,
			keyValue: KeyValue{
				Key:   "foo.bar[0].baz",
				Value: "zoo",
			},
			expectedResult: KeyValueEqual,
			expectedValues: nil,
			expectedError:  nil,
		},
		{
			name: "correct value - case sensitivity",
			node: singleValueNode,
			keyValue: KeyValue{
				Key:   "foo.bar[1].zap",
				Value: "bam",
			},
			expectedResult: KeyValueEqual,
			expectedValues: nil,
			expectedError:  nil,
		},
		{
			name: "optional value not present",
			node: singleValueNode,
			keyValue: KeyValue{
				Key:      "foo.bar[0].bazzle",
				Value:    "zoom",
				Optional: true,
			},
			expectedResult: KeyValueEqual,
			expectedValues: nil,
			expectedError:  nil,
		},
		{
			name: "multivalue - disallowed list not provided",
			node: multiValueNode,
			keyValue: KeyValue{
				Key:    "foo.bar",
				IsList: true,
			},
			expectedResult: KeyValueError,
			expectedValues: nil,
			expectedError:  errors.New("list of allowed or disallowed values not provided"),
		},
		{
			name: "multivalue - disallowed values in yaml",
			node: multiValueNode,
			keyValue: KeyValue{
				Key:        "foo.bar",
				IsList:     true,
				Disallowed: []string{"baz", "zoo"},
			},
			expectedResult: KeyValueDisallowedFound,
			expectedValues: []string{"baz", "zoo"},
			expectedError:  nil,
		},
		{
			name: "multivalue - no disallowed values in yaml",
			node: multiValueNode,
			keyValue: KeyValue{
				Key:        "foo.bar",
				IsList:     true,
				Disallowed: []string{"this should", "be a success"},
			},
			expectedResult: KeyValueEqual,
			expectedValues: nil,
			expectedError:  nil,
		},
		{
			name: "multivalue - allowed values in yaml all match",
			node: multiValueNode,
			keyValue: KeyValue{
				Key:     "foo.bar",
				IsList:  true,
				Allowed: []string{"baz", "zoo", "zoom"},
			},
			expectedResult: KeyValueEqual,
			expectedValues: nil,
			expectedError:  nil,
		},
		{
			name: "multivalue - value not in allowed list",
			node: multiValueNode,
			keyValue: KeyValue{
				Key:     "foo.bar",
				IsList:  true,
				Allowed: []string{"baz", "zoo"},
			},
			expectedResult: KeyValueDisallowedFound,
			expectedValues: []string{"zoom"},
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assert := assert.New(t)
			kvr, values, err := CheckKeyValue(tt.node, tt.keyValue)
			assert.Equal(tt.expectedResult, kvr, "expected result to match")
			assert.EqualValues(tt.expectedValues, values, "expected values to match")
			assert.Equal(err, tt.expectedError, "expected error to match")
		})
	}
}

func TestYamlBase(t *testing.T) {
	assert := assert.New(t)

	c := YamlBase{}
	c.HasData(true)
	assert.ElementsMatch([]breach.Breach{&breach.ValueBreach{
		BreachType: breach.BreachTypeValue,
		Value:      "no data available"}},
		c.Result.Breaches)

	mockCheck := func() YamlBase {
		return YamlBase{
			CheckBase: config.CheckBase{
				DataMap: map[string][]byte{
					"data": []byte(`
check:
  interval_days: 7
notification:
  emails:
    - admin@example.com
`),
				},
			},
			Values: []KeyValue{
				{
					Key:   "check.interval_days",
					Value: "7",
				},
			},
		}
	}

	c = mockCheck()
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Breaches))
	assert.EqualValues([]string{"[data] 'check.interval_days' equals '7'"}, c.Result.Passes)

	// Wrong key, correct value.
	c = mockCheck()
	c.Values = []KeyValue{
		{
			Key:   "check.interval",
			Value: "7",
		},
	}
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Passes))
	assert.ElementsMatch([]breach.Breach{&breach.KeyValueBreach{
		BreachType: breach.BreachTypeKeyValue,
		KeyLabel:   "config",
		Key:        "data",
		ValueLabel: "key not found",
		Value:      "check.interval"}},
		c.Result.Breaches)

	// Correct key, wrong value.
	c = mockCheck()
	c.Values = []KeyValue{
		{
			Key:   "check.interval_days",
			Value: "8",
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Passes))
	assert.EqualValues([]breach.Breach{&breach.KeyValueBreach{
		BreachType:    breach.BreachTypeKeyValue,
		KeyLabel:      "config:data",
		Key:           "check.interval_days",
		ValueLabel:    "actual",
		Value:         "7",
		ExpectedValue: "8"}},
		c.Result.Breaches)

	// Multiple config values - all correct.
	c = mockCheck()
	c.Values = []KeyValue{
		{
			Key:   "check.interval_days",
			Value: "7",
		},
		{
			Key:   "notification.emails[0]",
			Value: "admin@example.com",
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Breaches))
	assert.EqualValues(
		[]string{
			"[data] 'check.interval_days' equals '7'",
			"[data] 'notification.emails[0]' equals 'admin@example.com'",
		},
		c.Result.Passes)

	// Wildcard key.
	c = mockCheck()
	c.DataMap = map[string][]byte{
		"data": []byte(`
abcd:
  some:
    - thing 1
    - thing 2
    - thing 3
efgh:
  some:
    - thing 1
    - thing 2
    - thing 3
`),
	}
	c.Values = []KeyValue{
		{
			Key:        "*.some",
			IsList:     true,
			Disallowed: []string{"thing 2", "thing 4"},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Passes))
	assert.ElementsMatch([]breach.Breach{&breach.KeyValuesBreach{
		BreachType: breach.BreachTypeKeyValues,
		KeyLabel:   "config",
		Key:        "data",
		ValueLabel: "disallowed *.some",
		Values:     []string{"thing 2"}}},
		c.Result.Breaches)
}

func TestYamlBaseListValues(t *testing.T) {
	assert := assert.New(t)

	mockCheck := func() YamlBase {
		return YamlBase{
			CheckBase: config.CheckBase{
				DataMap: map[string][]byte{
					"data": []byte(`
foo:
  - a
  - b
  - c
  - d
`),
				},
			},
			Values: []KeyValue{
				{
					Key:        "foo",
					IsList:     true,
					Disallowed: []string{"b", "c"},
				},
			},
		}
	}
	c := mockCheck()
	c.UnmarshalDataMap()
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Passes))
	assert.ElementsMatch([]breach.Breach{&breach.KeyValuesBreach{
		BreachType: breach.BreachTypeKeyValues,
		KeyLabel:   "config",
		Key:        "data",
		ValueLabel: "disallowed foo",
		Values:     []string{"b", "c"}}},
		c.Result.Breaches)

	c = mockCheck()
	c.Values[0].Disallowed = []string{"e"}
	c.UnmarshalDataMap()
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Breaches))
	assert.EqualValues([]string{"[data] no disallowed 'foo'"}, c.Result.Passes)
}
