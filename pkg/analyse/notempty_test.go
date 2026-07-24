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
	plugin := Manager().GetFactories()["not:empty"]("testNotEmpty")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*NotEmpty)
	assert.True(ok)
	assert.Equal("testNotEmpty", analyser.Id)
}

func TestNotEmptyPluginName(t *testing.T) {
	instance := NewNotEmpty("testNotEmpty")
	assert.Equal(t, "not:empty", instance.GetName())
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
		{
			name: "listStringNil",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]string(nil),
			),
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "listStringEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]string{},
			),
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "listStringNotEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]string{"web/adminer.php", "web/foo.php"},
			),
			expectedBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "listStringNotEmpty",
					ValueLabel: "not empty",
					Value:      "web/adminer.php",
				},
				&breach.ValueBreach{
					BreachType: "value",
					CheckName:  "listStringNotEmpty",
					ValueLabel: "not empty",
					Value:      "web/foo.php",
				},
			},
		},
		{
			name: "mapStringEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]string{},
			),
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapStringNotEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]string{"adminer.php": "web/adminer.php"},
			),
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapStringNotEmpty",
					Key:        "adminer.php",
					ValueLabel: "not empty",
					Value:      "web/adminer.php",
				},
			},
		},
		{
			name: "mapBytesNotEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatMapBytes,
				map[string][]byte{"adminer.php": []byte("<?php")},
			),
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapBytesNotEmpty",
					Key:        "adminer.php",
					ValueLabel: "not empty",
					Value:      "<?php",
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
			analyser := NewNotEmpty(tc.name)

			tc.input.Collect()
			analyser.SetInput(tc.input)
			analyser.Analyse()

			assert.Len(analyser.Result.Breaches, len(tc.expectedBreaches))
			assert.ElementsMatch(tc.expectedBreaches, analyser.Result.Breaches)
		})
	}
}
