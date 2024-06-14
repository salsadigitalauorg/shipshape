package yaml

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"gopkg.in/yaml.v3"
)

// Key looks up a key in a YAML file using the file.lookup or
// yaml.key input plugins.
type Key struct {
	// Common fields.
	Name           string          `yaml:"name"`
	Format         data.DataFormat `yaml:"format"`
	ConnectionName string          `yaml:"connection"`
	InputName      string          `yaml:"input"`
	connection     connection.Connectioner
	input          fact.Facter
	errors         []error
	data           interface{}

	// Plugin fields.
	Path string `yaml:"path"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Key --package=yaml

func init() {
	fact.Registry["yaml.key"] = func(n string) fact.Facter { return &Key{Name: n} }
}

func (p *Key) PluginName() string {
	return "yaml.key"
}

func (p *Key) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Key) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportRequired, []string{"file.read", "file.lookup", "yaml.key"}
}

func (p *Key) Collect() {
	var yamlNodes []*yaml.Node
	var mapYamlNodes MapYamlNodes

	log.WithFields(log.Fields{
		"fact-plugin":  p.PluginName(),
		"fact":         p.Name,
		"input":        p.GetInputName(),
		"input-plugin": p.input.PluginName(),
	}).Debug("collecting data")

	switch p.input.PluginName() {
	case "file.read":
		inputData := data.AsBytes(p.input.GetData())
		if inputData == nil {
			return
		}

		var err error
		yamlNodes, err = PathLookupFromBytes(inputData, p.Path)
		if err != nil {
			p.errors = append(p.errors, err)
			return
		}
	case "file.lookup":
		inputData := data.AsMapStringBytes(p.input.GetData())
		if inputData == nil {
			return
		}

		var errs []error
		mapYamlNodes, errs = PathLookupFromMapBytes(inputData, p.Path)
		if len(errs) > 0 {
			p.errors = append(p.errors, errs...)
			return
		}
	case "yaml.key":
		switch p.input.GetFormat() {
		case FormatYamlNodes:
			yamlNodes = DataAsYamlNodes(p.input.GetData())
		case FormatMapYamlNodes:
			mapYamlNodes = DataAsMapYamlNodes(p.input.GetData())
		}
	}

	if yamlNodes == nil && mapYamlNodes == nil {
		return
	}

	switch p.Format {
	case FormatYamlNodes:
		if yamlNodes != nil {
			p.data = yamlNodes
		} else {
			p.errors = append(p.errors, errors.New("unsupported format for yaml-nodes key lookup"))
		}
	case FormatMapYamlKeys:
		if yamlNodes != nil {
			mappedData := MappingNodeToKeyedMap(yamlNodes[0])
			keys := make([]string, 0, len(mappedData))
			for k := range mappedData {
				keys = append(keys, k)
			}
			p.data = keys
		} else {
			p.errors = append(p.errors, errors.New("unsupported format for key lookup"))
		}
	case data.FormatMapString:
		if yamlNodes != nil {
			if len(yamlNodes) == 1 && yamlNodes[0].Kind == yaml.MappingNode {
				strMap, errs := YamlNodesToStringMapPathLookup(yamlNodes, p.Path)
				if len(errs) > 0 {
					p.errors = append(p.errors, errs...)
				}
				p.data = strMap
			}
		} else if mapYamlNodes != nil {
			p.data = mapYamlNodes.AsMapString()
		} else {
			p.errors = append(p.errors, errors.New("unsupported format for string key lookup"))
		}
	case FormatMapYamlNodes:
		if mapYamlNodes != nil {
			p.data = mapYamlNodes
		} else {
			p.errors = append(p.errors, errors.New("unsupported format for yaml-nodes-map key lookup"))
		}
	case data.FormatMapNestedString:
		if yamlNodes != nil {
			if len(yamlNodes) == 1 && yamlNodes[0].Kind == yaml.MappingNode {
				nestedStringMap, errs := YamlNodesToNestedStringMapPathLookup(yamlNodes, p.Path)
				if len(errs) > 0 {
					p.errors = append(p.errors, errs...)
				}
				p.data = nestedStringMap
			}
		} else if mapYamlNodes != nil {
			nestedStringMap := map[string]map[string]string{}
			for f, mapNodes := range mapYamlNodes {
				// Ensure the map is initialised.
				nestedStringMap[f] = nil
				if keyValue, errs := YamlNodesToStringMapPathLookup(mapNodes, p.Path); len(errs) > 0 {
					p.errors = append(p.errors, errs...)
				} else if len(keyValue) > 0 {
					nestedStringMap[f] = keyValue
				}
			}
			p.data = nestedStringMap
		} else {
			p.errors = append(p.errors, errors.New("unsupported format for nested-string-map key lookup"))
		}
	default:
		p.errors = append(p.errors, errors.New("unsupported format"))
	}
}
