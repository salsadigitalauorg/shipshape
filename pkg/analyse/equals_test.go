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

func TestEqualsInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the plugin is registered.
	plugin := Manager().GetFactories()["equals"]("TestEquals")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*Equals)
	assert.True(ok)
	assert.Equal("TestEquals", analyser.Id)
}

func TestEqualsPluginName(t *testing.T) {
	instance := NewEquals("TestEquals")
	assert.Equal(t, "equals", instance.GetName())
}

func TestEqualsAnalyse(t *testing.T) {
	tt := []internal.AnalyseTest{
		// String.
		{
			Name: "string",
			Input: testdata.New(
				"testFact",
				data.FormatString,
				"foo",
			),
			Analyser: &Equals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestEquals",
					},
					InputName: "testFact",
				},
				Value: "foo",
			},
			ExpectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "TestEquals",
					Value:      "testFact equals 'foo'",
				},
			},
		},
		{
			Name: "stringNotEqual",
			Input: testdata.New(
				"testFact",
				data.FormatString,
				"bar",
			),
			Analyser: &Equals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestEquals",
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
					"foo": "bar",
				},
			),
			Analyser: &Equals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestEquals",
					},
					InputName: "testFact",
				},
				Key:   "foo",
				Value: "bar",
			},
			ExpectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "TestEquals",
					Value:      "testFact equals 'bar'",
				},
			},
		},
		{
			Name: "mapStringNotEqual",
			Input: testdata.New(
				"testFact",
				data.FormatMapString,
				map[string]interface{}{"foo": "zoom"},
			),
			Analyser: &Equals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestEquals",
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
			Analyser: &Equals{
				BaseAnalyser: BaseAnalyser{
					BasePlugin: plugin.BasePlugin{
						Id: "TestEquals",
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
