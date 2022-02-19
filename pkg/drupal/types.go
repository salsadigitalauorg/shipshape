package drupal

import (
	"salsadigitalauorg/shipshape/pkg/core"
)

const (
	DBConfig      core.CheckType = "drupal-db-config"
	FileConfig    core.CheckType = "drupal-file-config"
	Modules       core.CheckType = "drupal-modules"
	ActiveModules core.CheckType = "drupal-active-modules"
)

type ConfigBase struct {
	core.CheckBase `yaml:",inline"`
	core.YamlCheck `yaml:",inline"`
	ConfigName     string `yaml:"config-name"`
	Pattern        string `yaml:"pattern"`
	ExcludePattern string `yaml:"exclude-pattern"`
}

type Drush struct {
	Alias   string `yaml:"alias"`
	Command string `yaml:"command"`
}

type FileConfigCheck struct {
	ConfigBase `yaml:",inline"`
	Path       string `yaml:"path"`
}

type DBConfigCheck struct {
	ConfigBase `yaml:",inline"`
	Drush      `yaml:",inline"`
}

type FileModuleCheck struct {
	FileConfigCheck `yaml:",inline"`
	Required        []string `yaml:"required"`
	Disallowed      []string `yaml:"disallowed"`
}

type ActiveModuleCheck struct {
	Drush          `yaml:",inline"`
	core.CheckBase `yaml:",inline"`
	core.YamlCheck `yaml:",inline"`
	Required       []string `yaml:"required"`
	Disallowed     []string `yaml:"disallowed"`
}
