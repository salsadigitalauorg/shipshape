package shipshape

import "gopkg.in/yaml.v3"

type CheckType string

const (
	Drush               CheckType = "Drush"
	DrupalDBConfig      CheckType = "DrupalDBConfig"
	DrupalFileConfig    CheckType = "DrupalFileConfig"
	DrupalModules       CheckType = "DrupalModules"
	DrupalActiveModules CheckType = "DrupalActiveModules"
)

type CheckList struct {
	Drush               []DrushCheck              `yaml:"drush,omitempty"`
	DrupalDBConfig      []DrupalDBConfigCheck     `yaml:"drupal-db-config,omitempty"`
	DrupalFileConfig    []DrupalFileConfigCheck   `yaml:"drupal-file-config,omitempty"`
	DrupalModules       []DrupalFileModuleCheck   `yaml:"drupal-modules,omitempty"`
	DrupalActiveModules []DrupalActiveModuleCheck `yaml:"drupal-active-modules,omitempty"`
}

type Config struct {
	DrupalRoot string    `yaml:"drupal-root"`
	Checks     CheckList `yaml:"checks"`
}

type Check interface {
	GetName() string
	FetchData() error
	RunCheck() (Result, error)
}

type CheckBase struct {
	Name   string `yaml:"name"`
	Data   []byte
	Result Result
}

type KeyValue struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type KeyValueResult int8

const (
	KeyValueError    KeyValueResult = -2
	KeyValueNotFound KeyValueResult = -1
	KeyValueNotEqual KeyValueResult = 0
	KeyValueEqual    KeyValueResult = 1
)

type YamlCheck struct {
	Values []KeyValue `yaml:"config-values"`
	Node   yaml.Node
}

type DrupalConfigBase struct {
	CheckBase  `yaml:",inline"`
	YamlCheck  `yaml:",inline"`
	ConfigName string `yaml:"config-name"`
}

type DrushCheck struct {
	Alias   string `yaml:"alias"`
	Command string `yaml:"command"`
}

type DrupalFileConfigCheck struct {
	DrupalConfigBase `yaml:",inline"`
	ConfigPath       string `yaml:"config-path"`
}

type DrupalDBConfigCheck struct {
	DrupalConfigBase `yaml:",inline"`
	DrushCheck       `yaml:",inline"`
}

type DrupalFileModuleCheck struct {
	DrupalFileConfigCheck `yaml:",inline"`
	Required              []string `yaml:"required"`
	Disallowed            []string `yaml:"disallowed"`
}

type DrupalActiveModuleCheck struct {
	DrushCheck `yaml:",inline"`
	Required   []string `yaml:"required"`
	Disallowed []string `yaml:"disallowed"`
}

type CheckStatus string

const (
	Pass CheckStatus = "Pass"
	Fail CheckStatus = "Fail"
)

type Result struct {
	CheckType CheckType
	Status    CheckStatus
	Passes    []string
	Failures  []string
}

type ResultList struct {
	Results []Result
	Errors  map[string]error
}
