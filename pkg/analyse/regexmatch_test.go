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

func TestRegexMatchInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the plugin is registered.
	plugin := Registry["regex:match"]("testRegexMatch")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*RegexMatch)
	assert.True(ok)
	assert.Equal("testRegexMatch", analyser.Id)
}

func TestRegexMatchPluginName(t *testing.T) {
	instance := RegexMatch{Id: "testRegexMatch"}
	assert.Equal(t, "regex:match", instance.PluginName())
}

func TestRegexMatchAnalyse(t *testing.T) {
	tt := []struct {
		name             string
		input            fact.Facter
		pattern          string
		ignore           string
		expectedBreaches []breach.Breach
	}{
		{
			name: "mapNestedStringEmpty",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapNestedString,
				TestInputData:       map[string]map[string]string{},
			},
		},
		{
			name: "mapNestedStringNoMatch",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapNestedString,
				TestInputData: map[string]map[string]string{
					"key1": {
						"subkey1": "value1",
						"subkey2": "value2",
					},
				},
			},
			pattern:          ".*value3.*",
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapNestedString1Match",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapNestedString,
				TestInputData: map[string]map[string]string{
					"key1": {
						"subkey1": "value1",
						"subkey2": "value2",
					},
				},
			},
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
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapNestedString,
				TestInputData: map[string]map[string]string{
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
			},
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
	}

	for _, tc := range tt {
		assert := assert.New(t)

		currLogOut := logrus.StandardLogger().Out
		defer logrus.SetOutput(currLogOut)
		logrus.SetOutput(io.Discard)

		t.Run(tc.name, func(t *testing.T) {
			analyser := RegexMatch{
				Id:      tc.name,
				Pattern: tc.pattern,
				Ignore:  tc.ignore,
			}

			tc.input.Collect()
			analyser.SetInput(tc.input)
			analyser.Analyse()

			assert.Len(analyser.Result.Breaches, len(tc.expectedBreaches))
			assert.ElementsMatch(tc.expectedBreaches, analyser.Result.Breaches)
		})
	}
}
