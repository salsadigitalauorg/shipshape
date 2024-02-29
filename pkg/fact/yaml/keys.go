package yaml

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"

	log "github.com/sirupsen/logrus"
)

// Keys looks up a key in a YAML file using the file.lookup or
// yaml.keys input plugins.
type Keys struct {
	// Common fields.
	Name           string          `yaml:"name"`
	Format         fact.FactFormat `yaml:"format"`
	ConnectionName string          `yaml:"connection"`
	InputName      string          `yaml:"input"`
	connection     connection.Connectioner
	input          fact.Facter
	errors         []error
	data           interface{}

	// Plugin fields.
	Path string `yaml:"path"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Keys --package=yaml

func init() {
	fact.Registry["yaml.keys"] = func(n string) fact.Facter { return &Keys{Name: n} }
}

func (p *Keys) PluginName() string {
	return "yaml.keys"
}

func (p *Keys) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Keys) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportRequired, []string{"file.lookup", "yaml.keys"}
}

func (p *Keys) Gather() {
	var lookupData data.MapYamlNodes
	var nestedLookupData map[string]data.MapYamlNodes
	var nestedStringMap map[string]map[string]string
	var errs []error

	log.WithFields(log.Fields{
		"fact-plugin":  p.PluginName(),
		"fact":         p.Name,
		"input":        p.GetInputName(),
		"input-plugin": p.input.PluginName(),
	}).Debug("gathering data")

	switch p.input.PluginName() {
	case "file.lookup":
		inputData := data.AsMapStringBytes(p.input.GetData())
		if inputData == nil {
			return
		}
		lookupData, errs = PathLookupFromBytes(inputData, p.Path)
	case "yaml.keys":
		inputData := data.AsMapYamlNodes(p.input.GetData())
		if inputData == nil {
			return
		}

		if p.Format == fact.FormatMapYamlNodes {
			nestedLookupData, errs = PathLookupFromYamlNodes(inputData, p.Path)
		} else if p.Format == fact.FormatMapNestedString {
			nestedStringMap = map[string]map[string]string{}
			for f, mapNodes := range inputData {
				// Ensure the map is initialised.
				nestedStringMap[f] = nil
				var keyValue map[string]string
				keyValue, errs = YamlMapNodesToStringMapPathLookup(mapNodes, p.Path)
				if len(keyValue) > 0 {
					nestedStringMap[f] = keyValue
				}
			}
		}
	}

	if len(errs) > 0 {
		p.errors = append(p.errors, errs...)
		return
	}

	if lookupData == nil && nestedLookupData == nil && nestedStringMap == nil {
		return
	}

	switch p.Format {
	case fact.FormatMapString:
		if lookupData != nil {
			p.data = lookupData.AsMapString()
		} else {
			p.errors = append(p.errors, errors.New("unsupported format for nested lookup"))
		}
	case fact.FormatMapYamlNodes:
		if lookupData != nil {
			p.data = lookupData
		} else {
			p.errors = append(p.errors, errors.New("unsupported format for nested lookup"))
		}
	case fact.FormatMapNestedString:
		if nestedStringMap != nil {
			p.data = nestedStringMap
		} else {
			p.errors = append(p.errors, errors.New("unsupported format for nested string map"))
		}
	default:
		p.errors = append(p.errors, errors.New("unsupported format"))
	}
}
