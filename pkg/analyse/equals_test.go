package analyse_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact/testdata"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
)

func TestEqualsInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the plugin is registered.
	plugin := Registry["equals"]("TestEquals")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*Equals)
	assert.True(ok)
	assert.Equal("TestEquals", analyser.Id)
}

func TestEqualsPluginName(t *testing.T) {
	instance := Equals{Id: "TestEquals"}
	assert.Equal(t, "equals", instance.PluginName())
}

func TestEqualsAnalyse(t *testing.T) {
	tt := []internal.AnalyseTest{
		// String.
		{
			Name: "string",
			Input: &testdata.TestFacter{
				Name:                "testFact",
				TestInputDataFormat: data.FormatString,
				TestInputData:       "foo",
			},
			Analyser: &Equals{
				InputName: "testFact",
				Id:        "TestEquals",
				Value:     "foo",
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
			Input: &testdata.TestFacter{
				Name:                "testFact",
				TestInputDataFormat: data.FormatString,
				TestInputData:       "bar",
			},
			Analyser: &Equals{
				InputName: "testFact",
				Id:        "TestEquals",
				Value:     "foo",
			},
			ExpectedBreaches: []breach.Breach{},
		},

		// Map of string.
		{
			Name: "mapString",
			Input: &testdata.TestFacter{
				Name:                "testFact",
				TestInputDataFormat: data.FormatMapString,
				TestInputData: map[string]interface{}{
					"foo": "bar",
				},
			},
			Analyser: &Equals{
				InputName: "testFact",
				Id:        "TestEquals",
				Key:       "foo",
				Value:     "bar",
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
			Input: &testdata.TestFacter{
				Name:                "testFact",
				TestInputDataFormat: data.FormatMapString,
				TestInputData:       map[string]interface{}{"foo": "zoom"},
			},
			Analyser: &Equals{
				InputName: "testFact",
				Id:        "TestEquals",
				Key:       "foo",
				Value:     "bar",
			},
			ExpectedBreaches: []breach.Breach{},
		},

		// Unsupported.
		{
			Name: "unsupported",
			Input: &testdata.TestFacter{
				Name:                "testFact",
				TestInputDataFormat: data.FormatListString,
				TestInputData:       []interface{}{"foo", "bar"},
			},
			Analyser: &Equals{
				InputName: "testFact",
				Id:        "TestEquals",
				Value:     "foo",
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
