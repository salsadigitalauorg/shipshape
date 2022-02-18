package drupal

import (
	"salsadigitalauorg/shipshape/pkg/core"
)

const (
	DrupalDBConfig      core.CheckType = "drupal-db-config"
	DrupalFileConfig    core.CheckType = "drupal-file-config"
	DrupalModules       core.CheckType = "drupal-modules"
	DrupalActiveModules core.CheckType = "drupal-active-modules"
)

type DrupalConfigBase struct {
	core.CheckBase `yaml:",inline"`
	core.YamlCheck `yaml:",inline"`
	ConfigName     string `yaml:"config-name"`
	ConfigPattern  string `yaml:"config-pattern"`
}

type Drush struct {
	Alias   string `yaml:"alias"`
	Command string `yaml:"command"`
}

type DrupalFileConfigCheck struct {
	DrupalConfigBase `yaml:",inline"`
	Path             string `yaml:"path"`
}

type DrupalDBConfigCheck struct {
	DrupalConfigBase `yaml:",inline"`
	Drush            `yaml:",inline"`
}

type DrupalFileModuleCheck struct {
	DrupalFileConfigCheck `yaml:",inline"`
	Required              []string `yaml:"required"`
	Disallowed            []string `yaml:"disallowed"`
}

type DrupalActiveModuleCheck struct {
	Drush          `yaml:",inline"`
	core.CheckBase `yaml:",inline"`
	core.YamlCheck `yaml:",inline"`
	Required       []string `yaml:"required"`
	Disallowed     []string `yaml:"disallowed"`
}
