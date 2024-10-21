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

func TestAllowedListInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the plugin is registered.
	plugin := Registry["allowed:list"]("testAllowedList")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*AllowedList)
	assert.True(ok)
	assert.Equal("testAllowedList", analyser.Id)
}

func TestAllowedListPluginName(t *testing.T) {
	instance := AllowedList{Id: "testAllowedList"}
	assert.Equal(t, "allowed:list", instance.PluginName())
}

func TestAllowedListAnalyse(t *testing.T) {
	tt := []struct {
		name             string
		input            fact.Facter
		allowed          []string
		deprecated       []string
		excludeKeys      []string
		ignore           []string
		expectedBreaches []breach.Breach
	}{
		{
			name: "mapStringNoBreaches",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapString,
				TestInputData: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			allowed:          []string{"value1", "value2"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapStringExcludedIgnored",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapString,
				TestInputData: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
					"key4": "value4",
				},
			},
			allowed:          []string{"value1", "value2"},
			excludeKeys:      []string{"key3"},
			ignore:           []string{"value4"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapStringNotAllowed",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapString,
				TestInputData: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
			allowed: []string{"value1", "value2"},
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "testAllowedList",
					KeyLabel:   "key",
					Key:        "key3",
					ValueLabel: "disallowed",
					Value:      "value3",
				},
			},
		},
		{
			name: "mapStringDeprecated",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapString,
				TestInputData: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
			allowed:    []string{"value1", "value2"},
			deprecated: []string{"value3"},
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "testAllowedList",
					KeyLabel:   "key",
					Key:        "key3",
					ValueLabel: "deprecated",
					Value:      "value3",
				},
			},
		},

		// Test data with map of list of strings.
		{
			name: "mapListStringNoBreaches",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapListString,
				TestInputData: map[string][]string{
					"key1": {"value1"},
					"key2": {"value2"},
				},
			},
			allowed:          []string{"value1", "value2"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapListStringExcludedIgnored",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapListString,
				TestInputData: map[string][]string{
					"key1": {"value1"},
					"key2": {"value2"},
					"key3": {"value3"},
					"key4": {"value4"},
				},
			},
			allowed:          []string{"value1", "value2"},
			excludeKeys:      []string{"key3"},
			ignore:           []string{"value4"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapListStringSomeIgnored",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapListString,
				TestInputData: map[string][]string{
					"key1": {"value1"},
					"key2": {"value2"},
					"key3": {"value3", "value4"},
				},
			},
			allowed: []string{"value1", "value2"},
			ignore:  []string{"value4"},
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "testAllowedList",
					KeyLabel:   "key",
					Key:        "key3",
					ValueLabel: "disallowed",
					Value:      "value3",
				},
			},
		},
		{
			name: "mapListStringDeprecated",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapListString,
				TestInputData: map[string][]string{
					"key1": {"value1"},
					"key2": {"value2", "value4"},
				},
			},
			allowed:    []string{"value1", "value2"},
			deprecated: []string{"value4"},
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "testAllowedList",
					KeyLabel:   "key",
					Key:        "key2",
					ValueLabel: "deprecated",
					Value:      "value4",
				},
			},
		},
		{
			name: "mapListStringDisallowed",
			input: &testdata.TestFacter{
				Name:                "testFacter",
				TestInputDataFormat: data.FormatMapListString,
				TestInputData: map[string][]string{
					"key1": {"value1", "value3"},
					"key2": {"value2", "value4"},
				},
			},
			allowed: []string{"value1", "value2"},
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "testAllowedList",
					KeyLabel:   "key",
					Key:        "key1",
					ValueLabel: "disallowed",
					Value:      "value3",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "testAllowedList",
					KeyLabel:   "key",
					Key:        "key2",
					ValueLabel: "disallowed",
					Value:      "value4",
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
			analyser := AllowedList{
				Id:          "testAllowedList",
				Allowed:     tc.allowed,
				Deprecated:  tc.deprecated,
				ExcludeKeys: tc.excludeKeys,
				Ignore:      tc.ignore,
			}

			tc.input.Collect()
			analyser.SetInput(tc.input)
			analyser.Analyse()

			assert.Len(analyser.Result.Breaches, len(tc.expectedBreaches))
			assert.ElementsMatch(tc.expectedBreaches, analyser.Result.Breaches)
		})
	}
}
