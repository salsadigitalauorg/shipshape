package analyse_test

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/fact/testdata"
)

const DataFormatUnsupported data.DataFormat = "nosupport"

func TestRegexMatchInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the plugin is registered.
	plugin := Manager().GetFactories()["regex:match"]("testRegexMatch")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*RegexMatch)
	assert.True(ok)
	assert.Equal("testRegexMatch", analyser.Id)
}

func TestRegexMatchPluginName(t *testing.T) {
	instance := NewRegexMatch("testRegexMatch")
	assert.Equal(t, "regex:match", instance.GetName())
}

func TestRegexMatchAnalyse(t *testing.T) {
	tt := []struct {
		name             string
		input            fact.Facter
		pattern          string
		expectedBreaches []breach.Breach
	}{
		{
			name: "nil",
			input: testdata.New(
				"testFacter",
				data.FormatNil,
				nil,
			),
		},

		// Nested string map.
		{
			name: "mapNestedStringEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatMapNestedString,
				map[string]map[string]string{},
			),
		},
		{
			name: "mapNestedStringNoMatch",
			input: testdata.New(
				"testFacter",
				data.FormatMapNestedString,
				map[string]map[string]string{
					"key1": {
						"subkey1": "value1",
						"subkey2": "value2",
					},
				},
			),
			pattern:          ".*value3.*",
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapNestedString1Match",
			input: testdata.New(
				"testFacter",
				data.FormatMapNestedString,
				map[string]map[string]string{
					"key1": {
						"subkey1": "value1",
						"subkey2": "value2",
					},
				},
			),
			pattern: ".*value2.*",
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedString1Match",
					Key:        "key1",
					ValueLabel: "subkey2",
					Value:      "value2",
				},
			},
		},
		{
			name: "mapNestedStringMultipleMatches",
			input: testdata.New(
				"testFacter",
				data.FormatMapNestedString,
				map[string]map[string]string{
					"key1": {
						"subkey1": "value1",
						"subkey2": "value2",
						"subkey4": "value4",
						"subkey5": "value5",
					},
					"key2": {
						"subkey2": "value2",
						"subkey3": "value3",
					},
				},
			),
			pattern: ".*value(1|3|5).*",
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedStringMultipleMatches",
					Key:        "key1",
					ValueLabel: "subkey1",
					Value:      "value1",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedStringMultipleMatches",
					Key:        "key1",
					ValueLabel: "subkey5",
					Value:      "value5",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedStringMultipleMatches",
					Key:        "key2",
					ValueLabel: "subkey3",
					Value:      "value3",
				},
			},
		},

		// String.
		{
			name: "stringEmpty/match/digit/single/match",
			input: testdata.New(
				"testFacter",
				data.FormatString,
				"0",
			),
			pattern: "^0$",
			expectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "stringEmpty/match/digit/single/match",
					Value:      " equals '0'",
				},
			},
		},
		{
			name: "stringEmpty/match/digit/single/nomatch",
			input: testdata.New(
				"testFacter",
				data.FormatString,
				"010",
			),
			pattern:          "^0$",
			expectedBreaches: []breach.Breach{},
		},

		// Unsupported.
		{
			name: "unsupported",
			input: testdata.New(
				"testFacter",
				DataFormatUnsupported,
				nil,
			),
			pattern: ".*",
			expectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "unsupported",
					Value:      "unsupported input format nosupport",
				},
			},
		},
	}

	for _, tc := range tt {
		currLogOut := logrus.StandardLogger().Out
		defer logrus.SetOutput(currLogOut)
		logrus.SetOutput(io.Discard)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			analyser := NewRegexMatch(tc.name)
			analyser.Pattern = tc.pattern

			tc.input.Collect()
			analyser.SetInput(tc.input)
			analyser.Analyse()

			assert.Len(analyser.Result.Breaches, len(tc.expectedBreaches))
			assert.ElementsMatch(tc.expectedBreaches, analyser.Result.Breaches)
		})
	}
}
