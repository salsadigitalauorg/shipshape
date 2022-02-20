package shipshape

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"

	"gopkg.in/yaml.v3"
)

func ReadAndParseConfig(projectDir string, f string) (Config, error) {
	cfg := Config{}
	data, err := ioutil.ReadFile(filepath.Join(projectDir, f))
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

func (cfg *Config) RunChecks() core.ResultList {
	rl := core.ResultList{Results: []core.Result{}}
	for _, checks := range cfg.Checks {
		for _, c := range checks {
			cfg.ProcessCheck(&rl, c)
		}
	}
	return rl
}

func (cfg *Config) ProcessCheck(rl *core.ResultList, c core.Check) {
	c.Init(cfg.ProjectDir, "")
	if c.RequiresData() {
		c.FetchData()
		c.HasData(true)
		c.UnmarshalDataMap()
	}
	if len(c.GetResult().Failures) == 0 {
		c.RunCheck()
	}
	rl.Results = append(rl.Results, c.GetResult())
}

func (cm *CheckMap) UnmarshalYAML(value *yaml.Node) error {
	newcm := make(CheckMap)
	for _, ct := range AllChecks {
		check_values, err := core.LookupYamlPath(value, string(ct))
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
			var c core.Check
			switch ct {
			case core.File:
				c = &core.FileCheck{}
			case core.Yaml:
				c = &core.YamlCheck{}
			case drupal.DrushYaml:
				c = &drupal.DrushYamlCheck{}
			case drupal.FileModule:
				c = &drupal.FileModuleCheck{}
			case drupal.DbModule:
				c = &drupal.DbModuleCheck{}
			default:
				continue
			}

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
