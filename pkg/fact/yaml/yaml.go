package yaml

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

func (m MapYamlNodes) AsMapString() map[string]string {
	newM := map[string]string{}
	for f, nodes := range m {
		for _, n := range nodes {
			newM[f] = n.Value
		}
	}
	return newM
}

func PathLookupFromBytes(fileData []byte, path string) ([]*yaml.Node, error) {
	n := yaml.Node{}
	err := yaml.Unmarshal(fileData, &n)
	if err != nil {
		log.WithError(err).Debug("failed to unmarshal yaml")
		return nil, err
	}

	foundNodes, err := utils.LookupYamlPath(&n, path)
	if err != nil {
		return nil, err
	}
	return foundNodes, nil
}

func PathLookupFromMapBytes(filesData map[string][]byte, path string) (MapYamlNodes, []error) {
	result := MapYamlNodes{}
	var errors []error
	for f, fBytes := range filesData {
		n := yaml.Node{}
		err := yaml.Unmarshal(fBytes, &n)
		if err != nil {
			errors = append(errors, err)
			log.WithError(err).Debug("failed to unmarshal yaml")
			continue
		}

		foundNodes, err := utils.LookupYamlPath(&n, path)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		result[f] = foundNodes
	}
	return result, errors
}

func PathLookupFromYamlNodes(filesNodes MapYamlNodes, path string) (map[string]MapYamlNodes, []error) {
	result := map[string]MapYamlNodes{}
	var errors []error
	// Each top-level node is expected to be a key.
	for f, nodes := range filesNodes {
		result[f] = MapYamlNodes{}
		for _, n := range nodes {
			foundNodes, err := utils.LookupYamlPath(n, path)
			if err != nil {
				errors = append(errors, err)
				log.WithError(err).Debug("failed to lookup yaml path from nodes")
				continue
			}
			result[f][n.Value] = foundNodes
		}
	}
	return result, errors
}

func YamlNodesToStringMapPathLookup(nodes []*yaml.Node, path string) (map[string]string, []error) {
	result := map[string]string{}
	var errs []error
	log.WithFields(log.Fields{
		"nodes": len(nodes),
		"path":  path,
	}).Debug("looking up yaml path from nodes")
	for _, mn := range nodes {
		mappedData := MappingNodeToKeyedMap(mn)
		for mapKey, n := range mappedData {
			foundNodes, err := utils.LookupYamlPath(n, path)
			if errs != nil {
				errs = append(errs, err)
				log.WithError(err).Debug("failed to lookup yaml path from nodes")
				continue
			}
			if len(foundNodes) == 0 {
				continue
			}
			result[mapKey] = foundNodes[0].Value
		}
	}
	return result, errs
}

func YamlNodesToNestedStringMapPathLookup(nodes []*yaml.Node, path string) (map[string]map[string]string, []error) {
	result := map[string]map[string]string{}
	var errs []error
	log.WithFields(log.Fields{
		"nodes": len(nodes),
		"path":  path,
	}).Debug("looking up yaml path from nodes")
	for _, mn := range nodes {
		mappedData := MappingNodeToKeyedMap(mn)
		for mapKey, n := range mappedData {
			foundNodes, err := utils.LookupYamlPath(n, path)
			if errs != nil {
				errs = append(errs, err)
				log.WithError(err).Debug("failed to lookup yaml path from nodes")
				continue
			}
			if len(foundNodes) == 0 {
				continue
			}
			for _, fn := range foundNodes {
				if fn.Kind != yaml.MappingNode {
					continue
				}
				result[mapKey] = MappingNodeToMapString(fn)
			}
		}
	}
	return result, errs

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
func MappingNodeToMapString(n *yaml.Node) map[string]string {
	result := map[string]string{}
	for i := 0; i < len(n.Content); i++ {
		if i%2 == 0 {
			result[n.Content[i].Value] = n.Content[i+1].Value
		}
	}
	return result
}

func DataAsYamlNodes(data interface{}) []*yaml.Node {
	if data == nil {
		return nil
	}
	return data.([]*yaml.Node)
}

func DataAsMapYamlNodes(data interface{}) MapYamlNodes {
	if data == nil {
		return nil
	}
	return data.(MapYamlNodes)
}
