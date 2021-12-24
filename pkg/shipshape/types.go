package shipshape

type CheckType string

const (
	DbConfig      CheckType = "DbConfig"
	FileConfig    CheckType = "FileConfig"
	Modules       CheckType = "Modules"
	DrushCommand  CheckType = "DrushCommand"
	ActiveModules CheckType = "ActiveModules"
)

type CheckList struct {
	DbConfig      []DbConfigCheck     `yaml:"db-config,omitempty"`
	FileConfig    []FileConfigCheck   `yaml:"file-config,omitempty"`
	Modules       []FileModuleCheck   `yaml:"modules,omitempty"`
	DrushCommand  []DrushCommandCheck `yaml:"drush-command,omitempty"`
	ActiveModules []ActiveModuleCheck `yaml:"active-modules,omitempty"`
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

type ConfigValue struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type DbConfigCheck struct {
	CheckBase
	ConfigName   string        `yaml:"config-name"`
	ConfigValues []ConfigValue `yaml:"config-values"`
}

type FileConfigCheck struct {
	CheckBase
	ConfigPath   string        `yaml:"config-path"`
	ConfigName   string        `yaml:"config-name"`
	ConfigValues []ConfigValue `yaml:"config-values"`
}

type FileModuleCheck struct {
	FileConfigCheck
	Required   []string `yaml:"required"`
	Disallowed []string `yaml:"disallowed"`
}

type DrushCommandCheck struct {
	CheckBase
	Alias   string `yaml:"alias"`
	Command string `yaml:"command"`
}

type ActiveModuleCheck struct {
	DrushCommandCheck
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
