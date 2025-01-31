package yaml

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"gopkg.in/yaml.v3"
)

// Key looks up a key in a YAML file using the file:lookup or
// yaml:key input plugins.
type Key struct {
	// Common fields.
	Name                 string          `yaml:"name"`
	Format               data.DataFormat `yaml:"format"`
	ConnectionName       string          `yaml:"connection"`
	InputName            string          `yaml:"input"`
	AdditionalInputNames []string        `yaml:"additional-inputs"`
	connection           connection.Connectioner
	input                fact.Facter
	additionalInputs     []fact.Facter
	errors               []error
	data                 interface{}

	// Resolve env vars.
	ResolveEnv bool   `yaml:"resolve-env"`
	EnvFile    string `yaml:"env-file"`

	// Plugin fields.
	Path string `yaml:"path"`
	// Only return the Yaml nodes found at the path.
	NodesOnly bool `yaml:"nodes-only"`
	// Only return the keys found at the path, if it's a map.
	KeysOnly bool `yaml:"keys-only"`
	// Ignore errors if the path is not found.
	IgnoreNotFound bool `yaml:"ignore-not-found"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Key --package=yaml --envresolver

func init() {
	fact.Registry["yaml:key"] = func(n string) fact.Facter { return &Key{Name: n} }
}

func (p *Key) PluginName() string {
	return "yaml:key"
}

func (p *Key) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Key) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportRequired, []string{"file:read", "file:lookup", "yaml:key"}
}

func (p *Key) Collect() {
	var lookup *YamlLookup
	var lookupMap *MapYamlLookup
	var nestedLookupMap map[string]*MapYamlLookup

	log.WithFields(log.Fields{
		"fact-plugin":  p.PluginName(),
		"fact":         p.Name,
		"input":        p.GetInputName(),
		"input-plugin": p.input.PluginName(),
	}).Debug("collecting data")

	switch p.input.GetFormat() {

	// The file:read plugin is used to read the file content.
	case data.FormatRaw:
		inputData := data.AsBytes(p.input.GetData())
		if inputData == nil {
			return
		}

		var err error
		lookup, err = NewYamlLookup(inputData, p.Path)
		if err != nil {
			if p.IgnoreNotFound && errors.Is(err, ErrPathNotFound) {
				p.Format = data.FormatNil
				return
			}
			p.errors = append(p.errors, err)
			return
		}

	// The file:lookup plugin is used to lookup files.
	case data.FormatMapBytes:
		inputData := data.AsMapBytes(p.input.GetData())
		if inputData == nil {
			return
		}

		var errs []error
		lookupMap, errs = NewMapYamlLookup(inputData, p.Path)
		if len(errs) > 0 {
			if p.IgnoreNotFound {
				allNotFound := true
				for _, err := range errs {
					if !errors.Is(err, ErrPathNotFound) {
						allNotFound = false
						break
					}
				}
				if allNotFound {
					p.Format = data.FormatNil
					return
				}
			}
			p.errors = append(p.errors, errs...)
			return
		}

	// The yaml:key plugin is used to lookup keys in a single YAML file.
	case FormatYamlNodes:
		yamlNodes := DataAsYamlNodes(p.input.GetData())
		var errs []error
		lookupMap, errs = NewMapYamlLookupFromNodes(yamlNodes, p.Path)
		if len(errs) > 0 {
			p.errors = append(p.errors, errs...)
			return
		}

	// The yaml.lookup plugin is used to lookup keys in multiple YAML files.
	case FormatMapYamlNodes:
		mapYamlNodes := DataAsMapYamlNodes(p.input.GetData())

		nestedLookupMap = map[string]*MapYamlLookup{}
		for f, nodes := range mapYamlNodes {
			lookupMap, errs := NewMapYamlLookupFromNodes(nodes, p.Path)
			if len(errs) > 0 {
				p.errors = append(p.errors, errs...)
				return
			}
			nestedLookupMap[f] = lookupMap
		}
	}

	if lookup == nil && lookupMap == nil && nestedLookupMap == nil {
		return
	}

	if p.NodesOnly {
		if lookup != nil {
			p.Format = FormatYamlNodes
			p.data = lookup.Nodes
		} else if lookupMap != nil {
			p.Format = FormatMapYamlNodes
			p.data = lookupMap.GetMapNodes()
		}
		return
	}

	if p.KeysOnly {
		if lookup != nil {
			if lookup.Kind != yaml.MappingNode {
				p.errors = append(p.errors, errors.New("keys-only lookup only supports a single mapping node"))
				return
			}
			mappedData := MappingNodeToKeyedMap(lookup.Nodes[0])
			keys := make([]string, 0, len(mappedData))
			for k := range mappedData {
				keys = append(keys, k)
			}
			p.Format = data.FormatListString
			p.data = keys
		} else {
			p.errors = append(p.errors, errors.New("yaml-nodes-map unsupported format for keys-only lookup"))
		}
		return
	}

	envMap, err := p.GetEnvMap()
	if err != nil {
		log.WithFields(log.Fields{
			"fact-plugin": p.PluginName(),
			"fact":        p.Name,
			"input":       p.GetInputName(),
		}).WithError(err).Error("unable to read env file")
		p.errors = append(p.errors, err)
		return
	}

	if lookup != nil {
		lookup.ProcessNodes(envMap)
		p.Format = lookup.Format
		p.data = lookup.Data
	} else if lookupMap != nil {
		lookupMap.ProcessMap(envMap)
		p.Format = lookupMap.Format
		p.data = lookupMap.DataMap
	} else {
		res := map[string]map[string]string{}
		for f, m := range nestedLookupMap {
			m.ProcessMap(envMap)
			if len(m.DataMap) == 0 {
				continue
			}
			res[f] = m.DataMapAsMapString()
			if p.Format == "" {
				switch m.Format {
				case data.FormatMapString:
					p.Format = data.FormatMapNestedString
				default:
					p.errors = append(p.errors, errors.New("unsupported format for nested lookup"))
					return
				}
			}
		}
		p.data = res
	}
}
