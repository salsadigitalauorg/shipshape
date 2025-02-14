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

func TestRegexNotMatchInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the plugin is registered.
	plugin := Registry["regex:not-match"]("testRegexNotMatch")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*RegexNotMatch)
	assert.True(ok)
	assert.Equal("testRegexNotMatch", analyser.Id)
}

func TestRegexNotMatchPluginName(t *testing.T) {
	instance := RegexNotMatch{Id: "testRegexNotMatch"}
	assert.Equal(t, "regex:not-match", instance.PluginName())
}

func TestRegexNotMatchAnalyse(t *testing.T) {
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
			name: "mapNestedStringAllMatch",
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
			pattern:          "value[12]",
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapNestedString1NotMatch",
			input: testdata.New(
				"testFacter",
				data.FormatMapNestedString,
				map[string]map[string]string{
					"key1": {
						"subkey1": "value1",
						"subkey2": "other2",
					},
				},
			),
			pattern: "value.*",
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedString1NotMatch",
					Key:        "key1",
					ValueLabel: "subkey2",
					Value:      "other2",
				},
			},
		},
		{
			name: "mapNestedStringMultipleNotMatches",
			input: testdata.New(
				"testFacter",
				data.FormatMapNestedString,
				map[string]map[string]string{
					"key1": {
						"subkey1": "other1",
						"subkey2": "value2",
						"subkey4": "other4",
						"subkey5": "value5",
					},
					"key2": {
						"subkey2": "other2",
						"subkey3": "other3",
					},
				},
			),
			pattern: "value.*",
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedStringMultipleNotMatches",
					Key:        "key1",
					ValueLabel: "subkey1",
					Value:      "other1",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedStringMultipleNotMatches",
					Key:        "key1",
					ValueLabel: "subkey4",
					Value:      "other4",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedStringMultipleNotMatches",
					Key:        "key2",
					ValueLabel: "subkey2",
					Value:      "other2",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedStringMultipleNotMatches",
					Key:        "key2",
					ValueLabel: "subkey3",
					Value:      "other3",
				},
			},
		},

		// String.
		{
			name: "string/notmatch/digit/single",
			input: testdata.New(
				"testFacter",
				data.FormatString,
				"1",
			),
			pattern: "^0$",
			expectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "string/notmatch/digit/single",
					Value:      " equals '1'",
				},
			},
		},
		{
			name: "string/match/digit/single",
			input: testdata.New(
				"testFacter",
				data.FormatString,
				"0",
			),
			pattern:          "^0$",
			expectedBreaches: []breach.Breach{},
		},

		// Unsupported.
		{
			name: "unsupported",
			input: testdata.New(
				"testFacter",
				data.FormatNil,
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
		assert := assert.New(t)

		currLogOut := logrus.StandardLogger().Out
		defer logrus.SetOutput(currLogOut)
		logrus.SetOutput(io.Discard)

		t.Run(tc.name, func(t *testing.T) {
			analyser := RegexNotMatch{
				Id:      tc.name,
				Pattern: tc.pattern,
			}

			tc.input.Collect()
			analyser.SetInput(tc.input)
			analyser.Analyse()

			assert.Len(analyser.Result.Breaches, len(tc.expectedBreaches))
			assert.ElementsMatch(tc.expectedBreaches, analyser.Result.Breaches)
		})
	}
}
