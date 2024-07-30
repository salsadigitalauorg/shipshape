package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
)

const (
	// FormatYamlNodes is used to represent []*yaml.Node.
	FormatYamlNodes data.DataFormat = "yaml-nodes"
	// FormatMapYamlNodes is used to represent map[string][]*yaml.Node.
	FormatMapYamlNodes data.DataFormat = "map-yaml-nodes"
)

type YamlLookup struct {
	Nodes  []*yaml.Node
	Path   string
	Kind   yaml.Kind
	Format data.DataFormat
	Data   interface{}
}

type MapYamlLookup struct {
	LookupMap map[string]*YamlLookup
	Path      string
	Format    data.DataFormat
	Kind      yaml.Kind
	DataMap   map[string]interface{}
}
