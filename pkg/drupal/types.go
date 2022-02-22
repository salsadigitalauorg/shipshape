package drupal

import (
	"salsadigitalauorg/shipshape/pkg/core"
)

const (
	DrushYaml  core.CheckType = "drush-yaml"
	FileModule core.CheckType = "drupal-file-module"
	DbModule   core.CheckType = "drupal-db-module"
)

type DrushCommand struct {
	DrushPath string `yaml:"drush-path"`
	Alias     string `yaml:"alias"`
	Command   string `yaml:"command"`
}

type DrushYamlCheck struct {
	core.YamlBase `yaml:",inline"`
	DrushCommand  `yaml:",inline"`
	ConfigName    string `yaml:"config-name"`
}

type FileModuleCheck struct {
	core.YamlCheck `yaml:",inline"`
	Required       []string `yaml:"required"`
	Disallowed     []string `yaml:"disallowed"`
}

type DbModuleCheck struct {
	DrushYamlCheck `yaml:",inline"`
	Required       []string `yaml:"required"`
	Disallowed     []string `yaml:"disallowed"`
}
