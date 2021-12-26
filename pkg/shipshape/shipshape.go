package shipshape

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
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
		Results: []Result{},
		Errors:  map[string]error{},
	}

	for _, checkType := range cfg.GetAllChecks() {
		checks, ok := checkType.([]Check)
		if !ok {
			continue
		}
		for _, c := range checks {
			r, err := c.RunCheck()
			if err != nil {
				rl.Errors[c.GetName()] = err
			}
			rl.Results = append(rl.Results, r)
		}
	}
	return rl
}

func (cfg *Config) GetAllChecks() []interface{} {
	return []interface{}{
		cfg.Checks.DrupalDBConfig,
		cfg.Checks.DrupalFileConfig,
		cfg.Checks.DrupalModules,
		cfg.Checks.Drush,
		cfg.Checks.DrupalActiveModules,
	}
}

func (c *CheckBase) GetName() string {
	return c.Name
}
