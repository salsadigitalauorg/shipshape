package analyse_test

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/analyse/testdata"
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestParseConfig(t *testing.T) {
	tt := []struct {
		name                string
		registry            map[string]func(string) Analyser
		config              map[string]map[string]interface{}
		expectAnalyserCount int
	}{
		{
			name:     "noPluginInRegistry",
			registry: nil,
			config: map[string]map[string]interface{}{
				"test": {
					"inexistentPlugin": map[string]interface{}{},
				},
			},
			expectAnalyserCount: 0,
		},
		{
			name: "pluginInRegistry",
			registry: map[string]func(string) Analyser{
				"test-analyser": func(id string) Analyser { return &testdata.TestAnalyser{} },
			},
			config: map[string]map[string]interface{}{
				"test": {
					"test-analyser": map[string]interface{}{},
				},
			},
			expectAnalyserCount: 1,
		},
	}

	for _, tc := range tt {
		currLogOut := logrus.StandardLogger().Out
		defer logrus.SetOutput(currLogOut)
		logrus.SetOutput(io.Discard)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Len(Manager().GetPlugins(), 0)
			factoriesBackup := Manager().GetFactories()
			if tc.registry != nil {
				Manager().Reset()
				for k, v := range tc.registry {
					Manager().RegisterFactory(k, v)
				}
			}
			Manager().ParseConfig(tc.config)
			defer func() {
				Manager().Reset()
				for k, v := range factoriesBackup {
					Manager().RegisterFactory(k, v)
				}
			}()
			assert.Len(Manager().GetPlugins(), tc.expectAnalyserCount)
		})
	}
}

func TestValidateInputs(t *testing.T) {
	tt := []struct {
		name             string
		analysers        map[string]Analyser
		expectErrorCount int
	}{
		{
			name:             "noAnalyser",
			analysers:        map[string]Analyser{},
			expectErrorCount: 0,
		},
		{
			name: "analyserWithNoError",
			analysers: map[string]Analyser{
				"test": &testdata.TestAnalyser{},
			},
			expectErrorCount: 0,
		},
		{
			name: "analyserWithError",
			analysers: map[string]Analyser{
				"test": &testdata.TestAnalyserInputError{},
			},
			expectErrorCount: 1,
		},
	}

	for _, tc := range tt {
		currLogOut := logrus.StandardLogger().Out
		defer logrus.SetOutput(currLogOut)
		logrus.SetOutput(io.Discard)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Len(Manager().GetErrors(), 0)

			if len(tc.analysers) > 0 {
				Manager().SetPlugins(tc.analysers)
			}

			Manager().ValidateInputs()

			defer func() {
				Manager().ResetPlugins()
				Manager().ResetErrors()
			}()
			assert.Len(Manager().GetErrors(), tc.expectErrorCount)
		})
	}
}

func TestAnalyseAll(t *testing.T) {
	tt := []struct {
		name          string
		analysers     map[string]Analyser
		expectResults map[string]result.Result
	}{
		{
			name:          "noAnalyser",
			analysers:     map[string]Analyser{},
			expectResults: map[string]result.Result{},
		},
		{
			name: "analyserWithPreProcessInputFail",
			analysers: map[string]Analyser{
				"test": &testdata.TestAnalyserPreprocessInputFail{
					BaseAnalyser: BaseAnalyser{
						BasePlugin: plugin.BasePlugin{
							Id: "test",
						},
					},
				},
			},
			expectResults: map[string]result.Result{
				"test": {
					Breaches: []breach.Breach{&breach.KeyValuesBreach{
						BreachType: "key-values",
						CheckName:  "test",
						Key:        "input failure",
						Values:     []string{"input error"},
					}},
				},
			},
		},
		{
			name: "analyserPass",
			analysers: map[string]Analyser{
				"test": &testdata.TestAnalyserPass{
					BaseAnalyser: BaseAnalyser{
						BasePlugin: plugin.BasePlugin{
							Id: "test",
						},
					},
				},
			},
			expectResults: map[string]result.Result{
				"test": {
					Breaches: []breach.Breach{&breach.KeyValuesBreach{
						BreachType: "key-values",
						CheckName:  "test",
						Key:        "breach found",
						Values:     []string{"more details would be here"},
					}},
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

			defer func() {
				Manager().ResetPlugins()
				Manager().ResetErrors()
			}()

			assert.Len(Manager().GetErrors(), 0)
			Manager().SetPlugins(tc.analysers)
			results := Manager().AnalyseAll()
			assert.Equal(tc.expectResults, results)
		})
	}
}
