package yaml

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/env"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

func NewYamlLookup(src []byte, path string) (*YamlLookup, error) {
	n := yaml.Node{}
	err := yaml.Unmarshal(src, &n)
	if err != nil {
		log.WithError(err).Debug("failed to unmarshal yaml")
		return nil, err
	}
	foundNodes, err := utils.LookupYamlPath(&n, path)
	if err != nil {
		log.WithError(err).Debug("failed to lookup yaml path")
		return nil, err
	}
	return &YamlLookup{Path: path, Nodes: foundNodes, Kind: foundNodes[0].Kind}, nil
}

func NewMapYamlLookup(src map[string][]byte, path string) (*MapYamlLookup, []error) {
	res := MapYamlLookup{Path: path, LookupMap: map[string]*YamlLookup{}}
	var errors []error
	for f, fBytes := range src {
		lookup, err := NewYamlLookup(fBytes, path)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if res.Kind == 0 {
			res.Kind = lookup.Kind
		}
		res.LookupMap[f] = lookup
	}
	return &res, errors
}

func NewMapYamlLookupFromNodes(nodes []*yaml.Node, path string) (*MapYamlLookup, []error) {
	res := MapYamlLookup{Path: path, LookupMap: map[string]*YamlLookup{}}
	var errs []error
	for _, n := range nodes {
		if n.Kind != yaml.MappingNode {
			errs = append(errs, errors.New("map-yaml-nodes lookup only supports mapping nodes"))
			continue
		}
		mappedNodes := MappingNodeToKeyedMap(n)
		for mapK, n := range mappedNodes {
			foundNodes, err := utils.LookupYamlPath(n, path)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			if foundNodes == nil {
				continue
			}

			res.LookupMap[mapK] = &YamlLookup{
				Path:  path,
				Nodes: foundNodes,
				Kind:  foundNodes[0].Kind,
			}
			if res.Kind == 0 {
				res.Kind = foundNodes[0].Kind
			}
		}
	}
	return &res, errs
}

func (y *YamlLookup) ProcessNodes(envMap map[string]string) {
	switch y.Kind {

	case yaml.ScalarNode:
		if y.Nodes[0].Value == "" {
			return
		}

		y.Format = data.FormatString
		resVal, err := env.ResolveValue(envMap, y.Nodes[0].Value)
		if err != nil {
			log.WithFields(log.Fields{
				"yaml-format": y.Format,
				"yaml-value":  y.Nodes[0].Value,
				"env-map":     envMap,
			}).WithError(err).Error("unable to resolve env var")
		}
		y.Data = resVal

	case yaml.SequenceNode:
		if y.Nodes[0].Content[0].Kind == yaml.ScalarNode {
			y.Format = data.FormatListString
			result := []string{}
			for _, n := range y.Nodes[0].Content {
				resVal, err := env.ResolveValue(envMap, n.Value)
				if err != nil {
					log.WithFields(log.Fields{
						"yaml-format": y.Format,
						"yaml-value":  n.Value,
						"env-map":     envMap,
					}).WithError(err).Error("unable to resolve env var")
					continue
				}
				result = append(result, resVal)
			}
			y.Data = result
		} else if y.Nodes[0].Content[0].Kind == yaml.MappingNode {
			y.Format = data.FormatListMapString
			result := []map[string]string{}
			for _, n := range y.Nodes[0].Content {
				result = append(result, MappingNodeToMapString(n, envMap))
			}
			y.Data = result
		}

	case yaml.MappingNode:
		y.Format = data.FormatMapString
		y.Data = MappingNodeToMapString(y.Nodes[0], envMap)

	case yaml.AliasNode:
		y.Format, y.Data = AliasNodeToData(y.Nodes[0].Alias, envMap)

	default:
		panic(fmt.Sprintf("unsupported kind '%d' in ProcessNodes", y.Kind))
	}
}

func (m *MapYamlLookup) GetMapNodes() map[string][]*yaml.Node {
	result := map[string][]*yaml.Node{}
	for f, lookup := range m.LookupMap {
		result[f] = lookup.Nodes
	}
	return result
}

func (m *MapYamlLookup) ProcessMap(envMap map[string]string) {
	m.DataMap = map[string]interface{}{}
	for f, lookup := range m.LookupMap {
		lookup.ProcessNodes(envMap)
		if lookup.Data == nil {
			continue
		}

		m.DataMap[f] = lookup.Data
		if m.Format == "" {
			switch lookup.Format {
			case data.FormatString:
				m.Format = data.FormatMapString
			case data.FormatListString:
				m.Format = data.FormatMapListString
			case data.FormatMapString:
				m.Format = data.FormatMapNestedString
			default:
				panic(fmt.Sprintf("unsupported format '%s' in ProcessMap", lookup.Format))
			}
		}
	}
}

func (m *MapYamlLookup) DataMapAsMapString() map[string]string {
	result := map[string]string{}
	for k, v := range m.DataMap {
		if vStr, ok := v.(string); ok {
			result[k] = vStr
		}
	}
	return result
}

// MappingNodeToKeyedMap converts a "mapping" yaml.Node to a keyed map.
func MappingNodeToKeyedMap(n *yaml.Node) map[string]*yaml.Node {
	result := map[string]*yaml.Node{}
	for i := 0; i < len(n.Content); i++ {
		if i%2 == 0 {
			result[n.Content[i].Value] = n.Content[i+1]
		}
	}
	return result
}

// MappingNodeToMapString converts a "mapping" yaml.Node to a map[string]string.
func MappingNodeToMapString(n *yaml.Node, envMap map[string]string) map[string]string {
	result := map[string]string{}
	for i := 0; i < len(n.Content); i++ {
		if i%2 == 0 {
			kNode := n.Content[i]
			vNode := n.Content[i+1]
			rawVal := vNode.Value
			if vNode.Kind == yaml.AliasNode {
				rawVal = vNode.Alias.Value
			}

			resVal, err := env.ResolveValue(envMap, rawVal)
			if err != nil {
				log.WithFields(log.Fields{
					"yaml-key":   kNode.Value,
					"yaml-value": rawVal,
					"env-map":    envMap,
				}).WithError(err).Error("unable to resolve env var")
				result[kNode.Value] = rawVal
				continue
			}
			result[kNode.Value] = resVal
		}
	}
	return result
}

func AliasNodeToData(n *yaml.Node, envMap map[string]string) (data.DataFormat, interface{}) {
	switch n.Kind {
	case yaml.ScalarNode:
		resVal, err := env.ResolveValue(envMap, n.Value)
		if err != nil {
			log.WithFields(log.Fields{
				"yaml-format": data.FormatString,
				"yaml-value":  n.Value,
				"env-map":     envMap,
			}).WithError(err).Warn("unable to resolve env var")
			return data.FormatString, n.Value
		}
		return data.FormatString, resVal
	case yaml.SequenceNode:
		if n.Content[0].Kind == yaml.ScalarNode {
			result := []string{}
			for _, n := range n.Content {
				resVal, err := env.ResolveValue(envMap, n.Value)
				if err != nil {
					log.WithFields(log.Fields{
						"yaml-format": data.FormatListString,
						"yaml-value":  n.Value,
						"env-map":     envMap,
					}).WithError(err).Warn("unable to resolve env var")
					result = append(result, n.Value)
					continue
				}
				result = append(result, resVal)
			}
			return data.FormatListString, result
		} else if n.Content[0].Kind == yaml.MappingNode {
			result := []map[string]string{}
			for _, n := range n.Content {
				result = append(result, MappingNodeToMapString(n, envMap))
			}
			return data.FormatListMapString, result
		}
	case yaml.MappingNode:
		return data.FormatMapString, MappingNodeToMapString(n, envMap)
	}
	panic(fmt.Sprintf("unsupported kind '%d' in AliasNodeToData", n.Kind))
}

func DataAsYamlNodes(data interface{}) []*yaml.Node {
	if data == nil {
		return nil
	}
	return data.([]*yaml.Node)
}

func DataAsMapYamlNodes(data interface{}) map[string][]*yaml.Node {
	if data == nil {
		return nil
	}
	return data.(map[string][]*yaml.Node)
}
