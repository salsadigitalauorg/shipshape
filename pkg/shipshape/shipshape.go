package shipshape

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func ReadAndParseConfig(f string) (Config, error) {
	c := Config{}
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return c, err
	}
	return ParseConfig(data)
}

func ParseConfig(data []byte) (Config, error) {
	c := Config{}
	err := yaml.Unmarshal(data, &c)
	return c, err
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

func (c *CheckBase) GetName() string {
	return c.Name
}

func (c *CheckBase) RunCheck() error {
	return nil
}

func (c *CheckBase) GetResult() Result {
	return c.Result
}

func (c *CheckBase) FetchData() error {
	return nil
}
