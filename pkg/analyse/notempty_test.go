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

func TestNotEmptyInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the plugin is registered.
	plugin := Registry["not:empty"]("testNotEmpty")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*NotEmpty)
	assert.True(ok)
	assert.Equal("testNotEmpty", analyser.Id)
}

func TestNotEmptyPluginName(t *testing.T) {
	instance := NotEmpty{Id: "testNotEmpty"}
	assert.Equal(t, "not:empty", instance.PluginName())
}

func TestNotEmptyAnalyse(t *testing.T) {
	tt := []struct {
		name             string
		input            fact.Facter
		expectedBreaches []breach.Breach
	}{
		{
			name: "mapNestedStringNil",
			input: testdata.New(
				"testFacter",
				data.FormatMapNestedString,
				map[string]map[string]string(nil),
			),
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapNestedStringEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatMapNestedString,
				map[string]map[string]string{},
			),
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapNestedStringNotEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatMapNestedString,
				map[string]map[string]string{
					"key1": {"subKey1": "value1"},
				},
			),
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapNestedStringNotEmpty",
					Key:        "key1",
					ValueLabel: "subKey1",
					Value:      "value1",
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
			analyser := NotEmpty{Id: tc.name}

			tc.input.Collect()
			analyser.SetInput(tc.input)
			analyser.Analyse()

			assert.Len(analyser.Result.Breaches, len(tc.expectedBreaches))
			assert.ElementsMatch(tc.expectedBreaches, analyser.Result.Breaches)
		})
	}
}
