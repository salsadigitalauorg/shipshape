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

func TestDriftInit(t *testing.T) {
	assert := assert.New(t)

	plugin := Manager().GetFactories()["drift"]("testDrift")
	assert.NotNil(plugin)
	analyser, ok := plugin.(*Drift)
	assert.True(ok)
	assert.Equal("testDrift", analyser.Id)
}

func TestDriftPluginName(t *testing.T) {
	instance := NewDrift("testDrift")
	assert.Equal(t, "drift", instance.GetName())
}

func TestDriftAnalyse(t *testing.T) {
	tt := []struct {
		name             string
		input            fact.Facter
		inputName        string
		expectedBreaches []breach.Breach
	}{
		{
			name: "listStringNil",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]string(nil),
			),
			inputName:        "ci-drift",
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "listStringEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]string{},
			),
			inputName:        "ci-drift",
			expectedBreaches: []breach.Breach{},
		},
		{
			name: "listStringNotEmpty",
			input: testdata.New(
				"testFacter",
				data.FormatListString,
				[]string{
					"--- template\n",
					"+++ current\n",
					"@@ -1,1 +1,1 @@\n",
					"-node-version: 20\n",
					"+node-version: 16\n",
				},
			),
			inputName: "ci-drift",
			expectedBreaches: []breach.Breach{
				&breach.KeyValuesBreach{
					BreachType: "key-values",
					CheckName:  "listStringNotEmpty",
					KeyLabel:   "file",
					Key:        "ci-drift",
					ValueLabel: "drift",
					Values: []string{
						"--- template\n",
						"+++ current\n",
						"@@ -1,1 +1,1 @@\n",
						"-node-version: 20\n",
						"+node-version: 16\n",
					},
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
			analyser := NewDrift(tc.name)
			analyser.InputName = tc.inputName

			tc.input.Collect()
			analyser.SetInput(tc.input)
			analyser.Analyse()

			assert.Len(analyser.Result.Breaches, len(tc.expectedBreaches))
			assert.ElementsMatch(tc.expectedBreaches, analyser.Result.Breaches)
		})
	}
}

// TestDriftAnalyseUnsupportedFormat ensures an unexpected input format does
// not panic and produces no breach, matching the defensive pattern used by
// other analysers (e.g. NotEmpty) for formats they cannot interpret.
func TestDriftAnalyseUnsupportedFormat(t *testing.T) {
	assert := assert.New(t)

	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	analyser := NewDrift("testDrift")
	input := testdata.New("testFacter", data.FormatString, "not-a-list")
	input.Collect()
	analyser.SetInput(input)

	assert.NotPanics(func() {
		analyser.Analyse()
	})
	assert.Empty(analyser.Result.Breaches)
}
