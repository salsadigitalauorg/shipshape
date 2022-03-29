// Package shipshape provides the basic types and base functions for parsing
// config, running checks as well as the file & yaml checks.
package shipshape

import (
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	"gopkg.in/yaml.v3"
)

func ReadAndParseConfig(projectDir string, f string) (Config, error) {
	cfg := Config{}
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return cfg, err
	}
	err = ParseConfig(data, projectDir, &cfg)
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

	return nil
}

func (cfg *Config) Init() {
	for ct, checks := range cfg.Checks {
		for _, c := range checks {
			c.Init(cfg.ProjectDir, ct)
		}
	}
}

func (cfg *Config) RunChecks(checkTypesToRun []string) ResultList {
	rl := ResultList{
		config:  cfg,
		Results: []Result{},
	}
	var wg sync.WaitGroup
	for ct, checks := range cfg.Checks {
		if len(checkTypesToRun) > 0 && !utils.StringSliceContains(checkTypesToRun, string(ct)) {
			continue
		}
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
	rl.IncrBreaches(c.GetResult().CheckType, len(c.GetResult().Failures))
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
