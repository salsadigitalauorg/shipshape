package analyse_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact/testdata"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

func TestNotEqualsInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the plugin is registered.
	plugin := Manager().GetFactories()["not:equals"]("TestNotEquals")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*NotEquals)
	assert.True(ok)
	assert.Equal("TestNotEquals", analyser.Id)
}

func TestNotEqualsPluginName(t *testing.T) {
	instance := NewNotEquals("TestNotEquals")
	assert.Equal(t, "not:equals", instance.GetName())
}

func TestNotEqualsAnalyse(t *testing.T) {
	tt := []internal.AnalyseTest{
		// String.
		{
			Name: "string",
			Input: testdata.New(
				"testFact",
				data.FormatString,
				"bar",
			),
			Analyser: &NotEquals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestNotEquals",
					},
					InputName: "testFact",
				},
				Value: "foo",
			},
			ExpectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "TestNotEquals",
					Value:      "testFact does not equal 'foo'",
				},
			},
		},
		{
			Name: "stringEqual",
			Input: testdata.New(
				"testFact",
				data.FormatString,
				"foo",
			),
			Analyser: &NotEquals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestNotEquals",
					},
					InputName: "testFact",
				},
				Value: "foo",
			},
			ExpectedBreaches: []breach.Breach{},
		},

		// Map of string.
		{
			Name: "mapString",
			Input: testdata.New(
				"testFact",
				data.FormatMapString,
				map[string]interface{}{
					"foo": "baz",
				},
			),
			Analyser: &NotEquals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestNotEquals",
					},
					InputName: "testFact",
				},
				Key:   "foo",
				Value: "bar",
			},
			ExpectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "TestNotEquals",
					Value:      "testFact does not equal 'bar'",
				},
			},
		},
		{
			Name: "mapStringEqual",
			Input: testdata.New(
				"testFact",
				data.FormatMapString,
				map[string]interface{}{"foo": "bar"},
			),
			Analyser: &NotEquals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestNotEquals",
					},
					InputName: "testFact",
				},
				Key:   "foo",
				Value: "bar",
			},
			ExpectedBreaches: []breach.Breach{},
		},

		// Unsupported.
		{
			Name: "unsupported",
			Input: testdata.New(
				"testFact",
				data.FormatListString,
				[]interface{}{"foo", "bar"},
			),
			Analyser: &NotEquals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestNotEquals",
					},
					InputName: "testFact",
				},
				Value: "foo",
			},
			ExpectedBreaches: []breach.Breach{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			internal.TestAnalyse(t, tc)
		})
	}
}
