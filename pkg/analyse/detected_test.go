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

func TestDetectedInit(t *testing.T) {
	assert := assert.New(t)

	plugin := Manager().GetFactories()["detected"]("testDetected")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*Detected)
	assert.True(ok)
	assert.Equal("testDetected", analyser.Id)
}

func TestDetectedPluginName(t *testing.T) {
	instance := NewDetected("testDetected")
	assert.Equal(t, "detected", instance.GetName())
}

func TestDetectedAnalyse(t *testing.T) {
	tt := []struct {
		name             string
		input            fact.Facter
		inputName        string
		expectedBreaches []breach.Breach
	}{
		{
			name: "mapStringNil",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]string(nil),
			),
			inputName:        "app-type",
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapStringEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]string{},
			),
			inputName:        "app-type",
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "mapStringSingle",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]string{
					"drupal": "15",
				},
			),
			inputName: "app-type",
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapStringSingle",
					KeyLabel:   "framework",
					Key:        "drupal",
					ValueLabel: "likelihood",
					Value:      "15",
				},
			},
		},
		{
			name: "mapStringMultiple",
			input: testdata.New(
				"testFacter",
				data.FormatMapString,
				map[string]string{
					"wordpress": "25",
					"drupal":    "15",
					"laravel":   "30",
				},
			),
			inputName: "app-type",
			expectedBreaches: []breach.Breach{
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapStringMultiple",
					KeyLabel:   "framework",
					Key:        "drupal",
					ValueLabel: "likelihood",
					Value:      "15",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapStringMultiple",
					KeyLabel:   "framework",
					Key:        "laravel",
					ValueLabel: "likelihood",
					Value:      "30",
				},
				&breach.KeyValueBreach{
					BreachType: "key-value",
					CheckName:  "mapStringMultiple",
					KeyLabel:   "framework",
					Key:        "wordpress",
					ValueLabel: "likelihood",
					Value:      "25",
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
			analyser := NewDetected(tc.name)
			analyser.InputName = tc.inputName

			tc.input.Collect()
			analyser.SetInput(tc.input)
			analyser.Analyse()

			assert.Len(analyser.Result.Breaches, len(tc.expectedBreaches))
			assert.ElementsMatch(tc.expectedBreaches, analyser.Result.Breaches)
		})
	}
}

// TestDetectedAnalyseBreachOrderDeterministic asserts that breach order is
// stable across repeated runs against the same multi-entry map input, since
// Go randomises map iteration and Detected must sort keys before emitting
// breaches to keep e2e assertions on breach order reliable.
func TestDetectedAnalyseBreachOrderDeterministic(t *testing.T) {
	assert := assert.New(t)

	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	inputData := map[string]string{
		"wordpress": "25",
		"drupal":    "15",
		"laravel":   "30",
		"symfony":   "20",
	}

	var firstOrder []string
	for i := 0; i < 25; i++ {
		input := testdata.New("testFacter", data.FormatMapString, inputData)
		input.Collect()

		analyser := NewDetected("testDetected")
		analyser.InputName = "app-type"
		analyser.SetInput(input)
		analyser.Analyse()

		order := make([]string, 0, len(analyser.Result.Breaches))
		for _, b := range analyser.Result.Breaches {
			kv, ok := b.(*breach.KeyValueBreach)
			assert.True(ok)
			order = append(order, kv.Key)
		}

		if firstOrder == nil {
			firstOrder = order
			continue
		}
		assert.Equal(firstOrder, order, "breach order changed between runs")
	}

	assert.Equal([]string{"drupal", "laravel", "symfony", "wordpress"}, firstOrder)
}

// TestDetectedAnalyseUnsupportedFormat ensures an unexpected input format
// does not panic and produces no breach, matching the defensive pattern
// used by other analysers (e.g. Drift, NotEmpty) for formats they cannot
// interpret.
func TestDetectedAnalyseUnsupportedFormat(t *testing.T) {
	assert := assert.New(t)

	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	analyser := NewDetected("testDetected")
	input := testdata.New("testFacter", data.FormatString, "not-a-map")
	input.Collect()
	analyser.SetInput(input)

	assert.NotPanics(func() {
		analyser.Analyse()
	})
	assert.Empty(analyser.Result.Breaches)
}
