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
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

func TestAllowedListInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the plugin is registered.
	plugin := Manager().GetFactories()["allowed:list"]("testAllowedList")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*AllowedList)
	assert.True(ok)
	assert.Equal("testAllowedList", analyser.Id)
}

func TestAllowedListPluginName(t *testing.T) {
	instance := NewAllowedList("testAllowedList")
	assert.Equal(t, "allowed:list", instance.GetName())
}

func TestAllowedListAnalyse(t *testing.T) {
	tt := []struct {
		name             string
		input            fact.Facter
		allowed          []string
		required         []string
		deprecated       []string
		excludeKeys      []string
		ignore           []string
		expectedBreaches []breach.Breach
	}{
		// List of strings.
		{
			name: "listString/NoBreaches",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]interface{}{"value1", "value2"},
			),
			allowed:          []string{"value1", "value2"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "listString/Ignored",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]interface{}{"value1", "value2", "value3"},
			),
			allowed:          []string{"value1", "value2"},
			ignore:           []string{"value3"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "listString/NotAllowed",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]interface{}{"value1", "value2", "value3"},
			),
			allowed: []string{"value1", "value2"},
			expectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "testAllowedList",
					ValueLabel: "disallowed value found",
					Value:      "value3",
				},
			},
		},
		{
			name: "listString/Deprecated",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]interface{}{"value1", "value2", "value3"},
			),
			allowed:    []string{"value1", "value2"},
			deprecated: []string{"value3"},
			expectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "testAllowedList",
					ValueLabel: "deprecated value found",
					Value:      "value3",
				},
			},
		},
		{
			name: "listString/Required",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]interface{}{"value1", "value2"},
			),
			allowed:  []string{"value1", "value2"},
			required: []string{"value3"},
			expectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "testAllowedList",
					ValueLabel: "required value not found",
					Value:      "value3",
				},
			},
		},

		// String map.
		{
			name: "mapStringNoBreaches",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			),
			allowed:          []string{"value1", "value2"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapStringExcludedIgnored",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
					"key4": "value4",
				},
			),
			allowed:          []string{"value1", "value2"},
			excludeKeys:      []string{"key3"},
			ignore:           []string{"value4"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapStringNotAllowed",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			),
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
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			),
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
		{
			name: "mapString/Required",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			),
			required: []string{"value4", "value5"},
			expectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "testAllowedList",
					ValueLabel: "required value not found",
					Value:      "value4",
				},
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "testAllowedList",
					ValueLabel: "required value not found",
					Value:      "value5",
				},
			},
		},

		// Test data with map of list of strings.
		{
			name: "mapListStringNoBreaches",
			input: testdata.New(
				"testFacter",
				data.FormatMapListString,
				map[string][]string{
					"key1": {"value1"},
					"key2": {"value2"},
				},
			),
			allowed:          []string{"value1", "value2"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapListStringExcludedIgnored",
			input: testdata.New(
				"testFacter",
				data.FormatMapListString,
				map[string][]string{
					"key1": {"value1"},
					"key2": {"value2"},
					"key3": {"value3"},
					"key4": {"value4"},
				},
			),
			allowed:          []string{"value1", "value2"},
			excludeKeys:      []string{"key3"},
			ignore:           []string{"value4"},
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapListStringSomeIgnored",
			input: testdata.New(
				"testFacter",
				data.FormatMapListString,
				map[string][]string{
					"key1": {"value1"},
					"key2": {"value2"},
					"key3": {"value3", "value4"},
				},
			),
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
			input: testdata.New(
				"testFacter",
				data.FormatMapListString,
				map[string][]string{
					"key1": {"value1"},
					"key2": {"value2", "value4"},
				},
			),
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
			input: testdata.New(
				"testFacter",
				data.FormatMapListString,
				map[string][]string{
					"key1": {"value1", "value3"},
					"key2": {"value2", "value4"},
				},
			),
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
		{
			name: "mapListString/Required",
			input: testdata.New(
				"testFacter",
				data.FormatMapListString,
				map[string][]string{
					"key1": {"value1", "value3", "value5"},
					"key2": {"value2", "value4"},
				},
			),
			required: []string{"value5", "value6"},
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "testAllowedList",
					Key:        "key1",
					ValueLabel: "required value not found",
					Value:      "value6",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "testAllowedList",
					Key:        "key2",
					ValueLabel: "required value not found",
					Value:      "value5",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "testAllowedList",
					Key:        "key2",
					ValueLabel: "required value not found",
					Value:      "value6",
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
			analyser := AllowedList{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "testAllowedList",
					},
				},
				Allowed:     tc.allowed,
				Required:    tc.required,
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
