package data

import "gopkg.in/yaml.v3"

type MapYamlNodes map[string][]*yaml.Node

func (m MapYamlNodes) AsMapString() map[string]string {
	newM := map[string]string{}
	for f, nodes := range m {
		for _, n := range nodes {
			newM[f] = n.Value
		}
	}
	return newM
}

func AsMapStringBytes(data interface{}) map[string][]byte {
	if data == nil {
		return nil
	}
	return data.(map[string][]byte)
}

func AsMapYamlNodes(data interface{}) MapYamlNodes {
	if data == nil {
		return nil
	}
	return data.(MapYamlNodes)
}

func AsNestedStringMap(data interface{}) map[string]map[string]string {
	if data == nil {
		return nil
	}
	return data.(map[string]map[string]string)
}
