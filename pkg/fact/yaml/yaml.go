package yaml

import (
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func PathLookupFromBytes(filesData map[string][]byte, path string) (data.MapYamlNodes, []error) {
	result := data.MapYamlNodes{}
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

func PathLookupFromYamlNodes(filesNodes data.MapYamlNodes, path string) (map[string]data.MapYamlNodes, []error) {
	result := map[string]data.MapYamlNodes{}
	var errors []error
	// Each top-level node is expected to be a key.
	for f, nodes := range filesNodes {
		result[f] = data.MapYamlNodes{}
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

func YamlMapNodesToStringMapPathLookup(mapNodes []*yaml.Node, path string) (map[string]string, []error) {
	result := map[string]string{}
	var errs []error
	log.WithFields(log.Fields{
		"mapNodes": len(mapNodes),
		"path":     path,
	}).Debug("looking up yaml path from map nodes")
	for _, mn := range mapNodes {
		mappedData := MappingNodeToKeyedMap(mn)
		for mapKey, n := range mappedData {
			foundNodes, err := utils.LookupYamlPath(n, path)
			if errs != nil {
				errs = append(errs, err)
				log.WithError(err).Debug("failed to lookup yaml path from nodes")
				continue
			}
			result[mapKey] = foundNodes[0].Value
		}
	}
	return result, errs
}

func MappingNodeToKeyedMap(n *yaml.Node) map[string]*yaml.Node {
	result := map[string]*yaml.Node{}
	for i := 0; i < len(n.Content); i++ {
		if i%2 == 0 {
			result[n.Content[i].Value] = n.Content[i+1]
		}
	}
	return result
}
