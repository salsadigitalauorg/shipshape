package shipshape

import (
	"errors"
	"io/ioutil"
	"os"
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/drupal"

	"gopkg.in/yaml.v3"
)

func ReadAndParseConfig(projectDir string, f string) (Config, error) {
	c := Config{}
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return c, err
	}
	err = ParseConfig(data, projectDir, &c)
	return c, err
}

func ParseConfig(data []byte, projectDir string, c *Config) error {
	err := yaml.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	if c.ProjectDir == "" && projectDir != "" {
		c.ProjectDir = projectDir
	} else {
		// Default project directory is current directory.
		projectDir, err = os.Getwd()
		if err != nil {
			return err
		}
		c.ProjectDir = projectDir
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
	rl := core.ResultList{
		Results: map[string]core.Result{},
		Errors:  map[string]error{},
	}

	for _, checks := range cfg.Checks {
		for _, c := range checks {
			cfg.ProcessCheck(&rl, c)
		}
	}
	return rl
}

func (cfg *Config) ProcessCheck(rl *core.ResultList, c core.Check) {
	c.Init(cfg.ProjectDir, "")
	err := c.FetchData()
	if err != nil {
		rl.Errors[c.GetName()] = err
	}

	err = c.RunCheck()
	if err != nil {
		rl.Errors[c.GetName()] = err
	}
	rl.Results[c.GetName()] = c.GetResult()
}

func (cm *CheckMap) UnmarshalYAML(value *yaml.Node) error {
	newcm := make(CheckMap)
	for _, ct := range AllChecks {
		check_values, err := core.LookupYamlPath(value, string(ct))
		if err != nil {
			return err
		}
		for _, cv := range check_values {
			var c core.Check
			switch ct {
			case drupal.DrupalDBConfig:
				c = &drupal.DrupalDBConfigCheck{}
			case drupal.DrupalFileConfig:
				c = &drupal.DrupalFileConfigCheck{}
			case drupal.DrupalModules:
				c = &drupal.DrupalFileModuleCheck{}
			case drupal.DrupalActiveModules:
				c = &drupal.DrupalActiveModuleCheck{}
			default:
				continue
			}

			if cv.Kind != yaml.SequenceNode {
				return errors.New("yaml: unmarshal errors")
			}
			err := cv.Content[0].Decode(c)
			if err != nil {
				return err
			}

			newcm[ct] = append(newcm[ct], c)
		}
	}
	*cm = newcm
	return nil
}
