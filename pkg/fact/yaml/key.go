package yaml

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/env"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

// Key looks up a key in a YAML file using the file:lookup or
// yaml:key input plugins.
type Key struct {
	fact.BaseFact       `yaml:",inline"`
	env.BaseEnvResolver `yaml:",inline"`

	// Plugin fields.
	Path string `yaml:"path"`
	// Only return the Yaml nodes found at the path.
	NodesOnly bool `yaml:"nodes-only"`
	// Only return the keys found at the path, if it's a map.
	KeysOnly bool `yaml:"keys-only"`
	// Ignore errors if the path is not found.
	IgnoreNotFound bool `yaml:"ignore-not-found"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --package=yaml

func init() {
	fact.Manager().RegisterFactory("yaml:key", func(n string) fact.Facter {
		return New(n)
	})
}

func New(id string) *Key {
	return &Key{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *Key) GetName() string {
	return "yaml:key"
}

func (p *Key) SupportedInputFormats() (plugin.SupportLevel, []data.DataFormat) {
	return plugin.SupportRequired, []data.DataFormat{
		data.FormatRaw,
		data.FormatMapBytes,
		FormatYamlNodes,
		FormatMapYamlNodes,
	}
}

func (p *Key) Collect() {
	var lookup *YamlLookup
	var lookupMap *MapYamlLookup
	var nestedLookupMap map[string]*MapYamlLookup

	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	contextLogger.WithFields(log.Fields{
		"input":        p.GetInputName(),
		"input-plugin": p.GetInput().GetName(),
		"input-format": p.GetInput().GetFormat(),
	}).Debug("collecting data")

	switch p.GetInput().GetFormat() {

	// The file:read plugin is used to read the file content.
	case data.FormatRaw:
		inputData := data.AsBytes(p.GetInput().GetData())
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
			contextLogger.WithError(err).Error("error looking up yaml path")
			p.AddErrors(err)
			return
		}

	// The file:lookup plugin is used to lookup files.
	case data.FormatMapBytes:
		inputData := data.AsMapBytes(p.GetInput().GetData())
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
			for _, err := range errs {
				contextLogger.WithError(err).Error("error looking up yaml path")
			}
			p.AddErrors(errs...)
			return
		}

	// The yaml:key plugin is used to lookup keys in a single YAML file.
	case FormatYamlNodes:
		yamlNodes := DataAsYamlNodes(p.GetInput().GetData())
		var errs []error
		lookupMap, errs = NewMapYamlLookupFromNodes(yamlNodes, p.Path)
		if len(errs) > 0 {
			for _, err := range errs {
				contextLogger.WithError(err).Error("error looking up yaml path")
			}
			p.AddErrors(errs...)
			return
		}

	// The yaml.lookup plugin is used to lookup keys in multiple YAML files.
	case FormatMapYamlNodes:
		mapYamlNodes := DataAsMapYamlNodes(p.GetInput().GetData())

		nestedLookupMap = map[string]*MapYamlLookup{}
		for f, nodes := range mapYamlNodes {
			lookupMap, errs := NewMapYamlLookupFromNodes(nodes, p.Path)
			if len(errs) > 0 {
				for _, err := range errs {
					contextLogger.WithError(err).Error("error looking up yaml path")
				}
				p.AddErrors(errs...)
				return
			}
			nestedLookupMap[f] = lookupMap
		}

	default:
		contextLogger.WithField("input-format", p.GetInput().GetFormat()).
			Error("unsupported input format")
	}

	if lookup == nil && lookupMap == nil && nestedLookupMap == nil {
		return
	}

	if p.NodesOnly {
		if lookup != nil {
			p.Format = FormatYamlNodes
			p.SetData(lookup.Nodes)
		} else if lookupMap != nil {
			p.Format = FormatMapYamlNodes
			p.SetData(lookupMap.GetMapNodes())
		}
		return
	}

	if p.KeysOnly {
		if lookup != nil {
			if lookup.Kind != yaml.MappingNode {
				contextLogger.Error("keys-only lookup only supports a single mapping node")
				p.AddErrors(errors.New("keys-only lookup only supports a single mapping node"))
				return
			}
			mappedData := MappingNodeToKeyedMap(lookup.Nodes[0])
			keys := make([]string, 0, len(mappedData))
			for k := range mappedData {
				keys = append(keys, k)
			}
			p.Format = data.FormatListString
			p.SetData(keys)
		} else {
			contextLogger.Error("yaml-nodes-map unsupported format for keys-only lookup")
			p.AddErrors(errors.New("yaml-nodes-map unsupported format for keys-only lookup"))
		}
		return
	}

	envMap, err := p.GetEnvMap()
	if err != nil {
		contextLogger.WithField("input", p.GetInputName()).
			WithError(err).Error("unable to read env file")
		p.AddErrors(err)
		return
	}

	if lookup != nil {
		lookup.ProcessNodes(envMap)
		p.Format = lookup.Format
		p.SetData(lookup.Data)
	} else if lookupMap != nil {
		lookupMap.ProcessMap(envMap)
		p.Format = lookupMap.Format
		p.SetData(lookupMap.DataMap)
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
					contextLogger.WithField("format", m.Format).
						Error("unsupported format for nested lookup")
					p.AddErrors(errors.New("unsupported format " + string(m.Format) + " for nested lookup"))
					return
				}
			}
		}
		p.SetData(res)
	}
}
