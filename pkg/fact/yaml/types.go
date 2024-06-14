package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
)

const (
	FormatYamlNodes    data.DataFormat = "yaml-nodes"
	FormatMapYamlKeys  data.DataFormat = "map-yaml-keys"
	FormatMapYamlNodes data.DataFormat = "map-yaml-nodes"
)

type MapYamlNodes map[string][]*yaml.Node
