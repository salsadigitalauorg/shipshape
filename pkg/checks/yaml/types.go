package yaml

import (
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"gopkg.in/yaml.v3"
)

// YamlBase represents the structure for a Yaml-based check.
type YamlBase struct {
	config.CheckBase `yaml:",inline"`
	Values           []KeyValue `yaml:"values"`
	Node             yaml.Node
	NodeMap          map[string]yaml.Node
}

// YamlCheck represents a Yaml file-based check, which can be for a single file
// or across a number of files defined by a regex pattern.
type YamlCheck struct {
	YamlBase       `yaml:",inline"`
	Path           string   `yaml:"path"`            // The directory in which to lookup files.
	File           string   `yaml:"file"`            // Single file name.
	Files          []string `yaml:"files"`           // A list of files to lint.
	Pattern        string   `yaml:"pattern"`         // Pattern-based files.
	ExcludePattern string   `yaml:"exclude-pattern"` // Pattern-based excluded files.

	// IgnoreMissing allows non-existent files to not be counted as a Fail.
	// Using a pointer here so we can differentiate between
	// false (default value) and an empty value.
	IgnoreMissing *bool `yaml:"ignore-missing"`
}

// YamlLintCheck represents a Yaml lint file-based check for a number of files.
type YamlLintCheck struct {
	YamlCheck `yaml:",inline"`
}

const (
	Yaml     config.CheckType = "yaml"     // Represents a YamlCheck.
	YamlLint config.CheckType = "yamllint" // Represents a YamlLintCheck.
)
