package drupal_test

import (
	"os/exec"
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"

	"github.com/stretchr/testify/assert"
)

func TestDrushYamlCheckInit(t *testing.T) {
	assert := assert.New(t)

	c := DrushYamlCheck{}
	c.Init(DrushYaml)
	assert.True(c.RequiresDb)
}

func TestDrushYamlMerge(t *testing.T) {
	assert := assert.New(t)

	c := DrushYamlCheck{
		YamlBase: yaml.YamlBase{
			Values: []yaml.KeyValue{
				{Key: "key1", Value: "val1", Optional: false},
			},
		},
		DrushCommand: DrushCommand{
			DrushPath: "/path/to/drush",
			Alias:     "alias1",
			Args:      []string{"arg1"},
		},
		Command:    "command1",
		ConfigName: "configname1",
	}
	c.Merge(&DrushYamlCheck{
		YamlBase: yaml.YamlBase{
			Values: []yaml.KeyValue{
				{Key: "key1", Value: "val1", Optional: true},
			},
		},
		DrushCommand: DrushCommand{
			DrushPath: "/new/path/to/drush",
			Alias:     "alias2",
			Args:      []string{"arg2"},
		},
		Command: "command2",
	})
	assert.EqualValues(DrushYamlCheck{
		YamlBase: yaml.YamlBase{
			Values: []yaml.KeyValue{
				{Key: "key1", Value: "val1", Optional: true},
			},
		},
		DrushCommand: DrushCommand{
			DrushPath: "/new/path/to/drush",
			Alias:     "alias2",
			Args:      []string{"arg2"},
		},
		Command:    "command2",
		ConfigName: "configname1",
	}, c)
}

func TestDrushYamlCheckFetchData(t *testing.T) {
	tt := []internal.FetchDataTest{
		{
			Name: "drushNotFound",
			Check: &DrushYamlCheck{
				Command:    "status",
				ConfigName: "core.extension",
			},
			ExpectBreaches: []result.Breach{&result.ValueBreach{
				BreachType: "value",
				CheckType:  "drush-yaml",
				Severity:   "normal",
				Value:      "vendor/drush/drush/drush: no such file or directory",
			}},
			ExpectStatusFail: true,
		},

		{
			Name: "drushError",
			Check: &DrushYamlCheck{
				Command:    "status",
				ConfigName: "core.extension",
			},
			PreFetch: func(t *testing.T) {
				command.ShellCommander = internal.ShellCommanderMaker(
					nil,
					&exec.ExitError{Stderr: []byte("unable to run drush command")},
					nil,
				)
			},
			ExpectBreaches: []result.Breach{&result.ValueBreach{
				BreachType: "value",
				CheckType:  "drush-yaml",
				Severity:   "normal",
				ValueLabel: "core.extension",
				Value:      "unable to run drush command",
			}},
			ExpectStatusFail: true,
		},

		{
			Name: "drushOK",
			Check: &DrushYamlCheck{
				Command:    "status",
				ConfigName: "core.extension",
			},
			PreFetch: func(t *testing.T) {
				stdout := `
module:
  block: 0
	views_ui: 0

`
				command.ShellCommander = internal.ShellCommanderMaker(
					&stdout,
					nil,
					nil,
				)
			},
		},
	}

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()

	for _, test := range tt {
		t.Run(test.Name, func(t *testing.T) {
			test.Check.Init(DrushYaml)
			internal.TestFetchData(t, test)
		})
	}
}

func TestDrushYamlCheckRunCheck(t *testing.T) {
	tests := []internal.RunCheckTest{
		{
			Name: "pass",
			Check: &DrushYamlCheck{
				YamlBase: yaml.YamlBase{
					CheckBase: config.CheckBase{
						DataMap: map[string][]byte{
							"core.extension": []byte(`{"profile":"standard"}`)},
					},
					Values: []yaml.KeyValue{
						{Key: "profile", Value: "standard"},
					},
				},
				ConfigName: "core.extension",
			},
			ExpectStatus: result.Pass,
			ExpectPasses: []string{"[core.extension] 'profile' equals 'standard'"},
		},
		{
			Name: "breach",
			Check: &DrushYamlCheck{
				YamlBase: yaml.YamlBase{
					CheckBase: config.CheckBase{
						DataMap: map[string][]byte{
							"core.extension": []byte(`{"profile":"minimal"}`)},
					},
					Values: []yaml.KeyValue{
						{Key: "profile", Value: "standard"},
					},
				},
				ConfigName: "core.extension",
			},
			PreRun: func(t *testing.T) {
				command.ShellCommander = internal.ShellCommanderMaker(nil, nil, nil)
			},
			ExpectFails: []result.Breach{&result.KeyValueBreach{
				BreachType:    "key-value",
				KeyLabel:      "core.extension",
				Key:           "profile",
				ValueLabel:    "actual",
				Value:         "minimal",
				ExpectedValue: "standard",
			}},
			ExpectStatus: result.Fail,
		},
		{
			Name: "breachMissingRemediation",
			Check: &DrushYamlCheck{
				YamlBase: yaml.YamlBase{
					CheckBase: config.CheckBase{
						DataMap: map[string][]byte{
							"core.extension": []byte(`{"profile":"minimal"}`)},
						PerformRemediation: true,
					},
					Values: []yaml.KeyValue{
						{Key: "profile", Value: "standard"},
					},
				},
				ConfigName: "core.extension",
			},
			PreRun: func(t *testing.T) {
				command.ShellCommander = internal.ShellCommanderMaker(nil, nil, nil)
			},
			ExpectFails: []result.Breach{&result.KeyValueBreach{
				BreachType:    "key-value",
				KeyLabel:      "core.extension",
				Key:           "profile",
				ValueLabel:    "actual",
				Value:         "minimal",
				ExpectedValue: "standard",
			}},
			ExpectStatus: result.Fail,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			curShellCommander := command.ShellCommander
			defer func() { command.ShellCommander = curShellCommander }()
			test.Check.UnmarshalDataMap()
			internal.TestRunCheck(t, test)
		})
	}
}
