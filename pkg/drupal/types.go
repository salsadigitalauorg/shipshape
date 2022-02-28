package drupal

import (
	"salsadigitalauorg/shipshape/pkg/shipshape"
)

const (
	DrushYaml     shipshape.CheckType = "drush-yaml"
	FileModule    shipshape.CheckType = "drupal-file-module"
	DbModule      shipshape.CheckType = "drupal-db-module"
	DbPermissions shipshape.CheckType = "drupal-db-permissions"
)

type DrushCommand struct {
	DrushPath string `yaml:"drush-path"`
	Alias     string `yaml:"alias"`
	Command   string `yaml:"command"`
}

type DrushYamlCheck struct {
	shipshape.YamlBase `yaml:",inline"`
	DrushCommand       `yaml:",inline"`
	ConfigName         string `yaml:"config-name"`
}

type FileModuleCheck struct {
	shipshape.YamlCheck `yaml:",inline"`
	Required            []string `yaml:"required"`
	Disallowed          []string `yaml:"disallowed"`
}

type DbModuleCheck struct {
	DrushYamlCheck `yaml:",inline"`
	Required       []string `yaml:"required"`
	Disallowed     []string `yaml:"disallowed"`
}

type DrushPermissions map[string]struct {
	Label string   `yaml:"label"`
	Perms []string `yaml:"perms"`
}

type DbPermissionsCheck struct {
	DrushYamlCheck `yaml:",inline"`
	Disallowed     []string `yaml:"disallowed"`
	Permissions    DrushPermissions
}
