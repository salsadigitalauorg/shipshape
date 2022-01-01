package shipshape

import (
	"errors"
	"io/ioutil"
	"os"

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

func (cfg *Config) RunChecks() ResultList {
	rl := ResultList{
		Results: map[string]Result{},
		Errors:  map[string]error{},
	}

	for _, checks := range cfg.Checks {
		for _, c := range checks {
			cfg.ProcessCheck(&rl, c)
		}
	}
	return rl
}

func (cfg *Config) ProcessCheck(rl *ResultList, c Check) {
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
		check_values, err := LookupYamlPath(value, string(ct))
		if err != nil {
			return err
		}
		for _, cv := range check_values {
			var c Check
			switch ct {
			case DrupalDBConfig:
				c = &DrupalDBConfigCheck{}
			case DrupalFileConfig:
				c = &DrupalFileConfigCheck{}
			case DrupalModules:
				c = &DrupalFileModuleCheck{}
			case DrupalActiveModules:
				c = &DrupalActiveModuleCheck{}
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

func (c *CheckBase) Init(pd string, ct CheckType) {
	c.ProjectDir = pd
	if c.Result.CheckType == "" {
		c.Result = Result{CheckType: ct}
	}
}

func (c *CheckBase) GetName() string {
	return c.Name
}

func (c *CheckBase) FetchData() error {
	return nil
}

func (c *CheckBase) RunCheck() error {
	return nil
}

func (c *CheckBase) GetResult() Result {
	return c.Result
}
