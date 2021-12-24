package shipshape

type CheckType string

const (
	ActiveConfig  CheckType = "ActiveConfig"
	ActiveModules CheckType = "ActiveModules"
	FileConfig    CheckType = "FileConfig"
	Modules       CheckType = "Modules"
)

type Result struct {
	CheckType CheckType
}

type Check interface {
	RunCheck() error
	GetResult() Result
}

type CheckBase struct {
	Name string `yaml:"name"`
}

type Config struct {
	DrupalRoot string    `yaml:"drupal-root"`
	Checks     CheckList `yaml:"checks"`
}

type CheckList struct {
	ActiveConfig  []ActiveConfigCheck `yaml:"active-config,omitempty"`
	ActiveModules []ActiveModuleCheck `yaml:"active-modules,omitempty"`
	FileConfig    []FileConfigCheck   `yaml:"file-config,omitempty"`
	Modules       []FileModuleCheck   `yaml:"modules,omitempty"`
}

type ActiveConfigCheck struct {
	CheckBase
	ConfigName string `yaml:"config-name"`
}

type ActiveModuleCheck struct {
	ActiveConfigCheck
	Required   []string `yaml:"required"`
	Disallowed []string `yaml:"disallowed"`
}

type FileConfigCheck struct {
	CheckBase
	ConfigPath string `yaml:"config-path"`
	ConfigName string `yaml:"config-name"`
}

type FileModuleCheck struct {
	FileConfigCheck
	Required   []string `yaml:"required"`
	Disallowed []string `yaml:"disallowed"`
}
