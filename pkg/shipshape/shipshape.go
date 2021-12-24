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

func (cfg *Config) RunChecks() {
	for _, c := range cfg.Checks.ActiveConfig {
		c.RunCheck()
	}
	for _, c := range cfg.Checks.ActiveModules {
		c.RunCheck()
	}
	for _, c := range cfg.Checks.FileConfig {
		c.RunCheck()
	}
	for _, c := range cfg.Checks.Modules {
		c.RunCheck()
	}
}
