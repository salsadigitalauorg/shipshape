package drupal

import (
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

const (
	DrushYaml     config.CheckType = "drush-yaml"
	FileModule    config.CheckType = "drupal-file-module"
	DbModule      config.CheckType = "drupal-db-module"
	DbPermissions config.CheckType = "drupal-db-permissions"

	DbUserTfa    config.CheckType = "drupal-db-user-tfa"
	TrackingCode config.CheckType = "drupal-tracking-code"
)

type DrushCommand struct {
	DrushPath string `yaml:"drush-path"`
	Alias     string `yaml:"alias"`
	Args      []string
}

type DrushYamlCheck struct {
	yaml.YamlBase    `yaml:",inline"`
	DrushCommand     `yaml:",inline"`
	Command          string `yaml:"command"`
	ConfigName       string `yaml:"config-name"`
	RemediateCommand string `yaml:"remediate-command"`
	RemediateMsg     string `yaml:"remediate-msg"`
}

type FileModuleCheck struct {
	yaml.YamlCheck `yaml:",inline"`
	Required       []string `yaml:"required"`
	Disallowed     []string `yaml:"disallowed"`
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
