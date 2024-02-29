package fact

import "gopkg.in/yaml.v3"

type DataMapYamlNodes map[string][]*yaml.Node

func (m DataMapYamlNodes) AsMapString() map[string]string {
	newM := map[string]string{}
	for f, nodes := range m {
		for _, n := range nodes {
			newM[f] = n.Value
		}
	}
	return newM
}

func DataAsMapStringBytes(data interface{}) map[string][]byte {
	if data == nil {
		return nil
	}
	return data.(map[string][]byte)
}

func DataAsMapYamlNodes(data interface{}) DataMapYamlNodes {
	if data == nil {
		return nil
	}
	return data.(DataMapYamlNodes)
}
