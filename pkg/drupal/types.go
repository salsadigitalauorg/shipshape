package drupal

import (
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

const (
	DrushYaml     shipshape.CheckType = "drush-yaml"
	FileModule    shipshape.CheckType = "drupal-file-module"
	DbModule      shipshape.CheckType = "drupal-db-module"
	DbPermissions shipshape.CheckType = "drupal-db-permissions"
	TrackingCode  shipshape.CheckType = "drupal-tracking-code"
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

type DrushRole struct {
	Label string   `yaml:"label"`
	Perms []string `yaml:"perms"`
}

type DbPermissionsCheck struct {
	DrushYamlCheck `yaml:",inline"`
	Disallowed     []string `yaml:"disallowed"`
	ExcludeRoles   []string `yaml:"exclude-roles"`
	Permissions    map[string]DrushRole
}

type DrushStatus struct {
	Uri string `yaml:"uri"`
}

type TrackingCodeCheck struct {
	DrushYamlCheck `yaml:",inline"`
	Code           string `yaml:"code"`
	DrushStatus    DrushStatus
}
