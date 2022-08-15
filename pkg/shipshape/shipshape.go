// Package shipshape provides the basic types and base functions for parsing
// config, running checks as well as the file & yaml checks.
package shipshape

import (
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/salsadigitalauorg/shipshape/pkg/merger"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	"gopkg.in/yaml.v3"
)

func ReadAndParseConfig(projectDir string, files []string) (Config, error) {

	cfg := Config{}
	merger := merger.NewMerger()

	var data []byte
	var err error

	for _, f := range files {

		if utils.StringIsUrl(f) {
			data, err = utils.FetchContentFromUrl(f)
			if err != nil {
				return cfg, err
			}
		} else {
			data, err = ioutil.ReadFile(f)
			if err != nil {
				return cfg, err
			}
		}

		err = merger.AddData(data)
		if err != nil {
			panic(err)
		}
	}

	err = merger.Save("/tmp/merged.yaml")
	if err != nil {
		panic(err)
	}

	err = ParseConfig(data, projectDir, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, err
}

func ParseConfig(data []byte, projectDir string, cfg *Config) error {
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	if cfg.ProjectDir == "" && projectDir != "" {
		cfg.ProjectDir = projectDir
	} else {
		// Default project directory is current directory.
		projectDir, _ = os.Getwd()
		cfg.ProjectDir = projectDir
	}

	if cfg.FailSeverity == "" {
		cfg.FailSeverity = HighSeverity
	}

	return nil
}

func (cfg *Config) Init() {
	for ct, checks := range cfg.Checks {
		for _, c := range checks {
			c.Init(cfg.ProjectDir, ct)
		}
	}
}

// FilterChecksToRun iterates over all the checks and filters them based on
// a provided list of check types to run or whether to exclude database checks.
func (cfg *Config) FilterChecksToRun(checkTypesToRun []string, excludeDb bool) {
	newCm := CheckMap{}
	for ct, checks := range cfg.Checks {
		newChecks := []Check{}
		for _, c := range checks {
			if len(checkTypesToRun) > 0 && !utils.StringSliceContains(checkTypesToRun, string(ct)) {
				continue
			}
			if excludeDb && c.RequiresDatabase() {
				continue
			}
			newChecks = append(newChecks, c)
		}
		if len(newChecks) > 0 {
			newCm[ct] = newChecks
		}
	}
	cfg.Checks = newCm
}

func (cfg *Config) RunChecks() ResultList {
	rl := ResultList{
		config:  cfg,
		Results: []Result{},
	}
	var wg sync.WaitGroup
	for ct, checks := range cfg.Checks {
		checks := checks
		rl.IncrChecks(ct, len(checks))
		for i := range checks {
			wg.Add(1)
			check := checks[i]
			go func() {
				defer wg.Done()
				cfg.ProcessCheck(&rl, check)
			}()
		}
	}
	wg.Wait()
	rl.Sort()
	return rl
}

func (cfg *Config) ProcessCheck(rl *ResultList, c Check) {
	c.Init(cfg.ProjectDir, "")
	if c.RequiresData() {
		c.FetchData()
		c.HasData(true)
		if len(c.GetResult().Failures) == 0 {
			c.UnmarshalDataMap()
		}
	}
	if len(c.GetResult().Failures) == 0 && len(c.GetResult().Passes) == 0 {
		c.RunCheck()
		c.GetResult().Sort()
	}
	rl.Results = append(rl.Results, *c.GetResult())
	rl.IncrBreaches(c, len(c.GetResult().Failures))
}

func (cm *CheckMap) UnmarshalYAML(value *yaml.Node) error {
	newcm := make(CheckMap)
	for ct, cFunc := range ChecksRegistry {
		check_values, err := LookupYamlPath(value, string(ct))
		if err != nil {
			return err
		}

		if len(check_values) == 0 {
			continue
		}

		if check_values[0].Kind != yaml.SequenceNode {
			return errors.New("yaml: unmarshal errors")
		}

		for _, cv := range check_values[0].Content {
			c := cFunc()
			err := cv.Decode(c)
			if err != nil {
				return err
			}
			newcm[ct] = append(newcm[ct], c)
		}
	}
	*cm = newcm
	return nil
}
