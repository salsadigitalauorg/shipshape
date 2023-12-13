package json

import (
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

// JsonCheck represents a JSON file-based check, which can be for a single file
// or across a number of files defined by a regex pattern.
type JsonCheck struct {
	yaml.YamlCheck `yaml:",inline"`
	// Cannot override struct field with same YAML key.
	// https://github.com/go-yaml/yaml/issues/467
	KeyValues []KeyValue     `yaml:"key-values"`
	Node      map[string]any `yaml:"-"`
}

const (
	Json config.CheckType = "json" // Represents a JsonCheck.
)
