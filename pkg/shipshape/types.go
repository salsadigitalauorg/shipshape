package shipshape

import "gopkg.in/yaml.v3"

type CheckType string

const (
	DrupalDBConfig      CheckType = "drupal-db-config"
	DrupalFileConfig    CheckType = "drupal-file-config"
	DrupalModules       CheckType = "drupal-modules"
	DrupalActiveModules CheckType = "drupal-active-modules"
)

var AllChecks = []CheckType{
	DrupalDBConfig,
	DrupalFileConfig,
	DrupalModules,
	DrupalActiveModules,
}

type CheckMap map[CheckType][]Check

type Config struct {
	ProjectDir string   `yaml:"project-dir"`
	DrupalRoot string   `yaml:"drupal-root"`
	Checks     CheckMap `yaml:"checks"`
}

type Check interface {
	Init(pd string, ct CheckType)
	GetName() string
	FetchData() error
	RunCheck() error
	GetResult() Result
}

type CheckBase struct {
	Name       string `yaml:"name"`
	ProjectDir string `yaml:""`
	Data       []byte
	Result     Result
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

type Drush struct {
	Alias   string `yaml:"alias"`
	Command string `yaml:"command"`
}

type DrupalFileConfigCheck struct {
	DrupalConfigBase `yaml:",inline"`
	ConfigPath       string `yaml:"config-path"`
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
	Drush      `yaml:",inline"`
	CheckBase  `yaml:",inline"`
	YamlCheck  `yaml:",inline"`
	Required   []string `yaml:"required"`
	Disallowed []string `yaml:"disallowed"`
}

type CheckStatus string

const (
	Pass CheckStatus = "Pass"
	Fail CheckStatus = "Fail"
)

type Result struct {
	Name      string      `json:"name"`
	CheckType CheckType   `json:"check-type"`
	Status    CheckStatus `json:"status"`
	Passes    []string    `json:"passes"`
	Failures  []string    `json:"failures"`
	Error     string      `json:"error"`
}

type ResultList struct {
	Results map[string]Result `json:"results"`
	Errors  map[string]error  `json:"errors"`
}
