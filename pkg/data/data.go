package data

import "gopkg.in/yaml.v3"

type DataFormat string

const (
	FormatRaw             DataFormat = "raw"
	FormatList            DataFormat = "list"
	FormatMapBytes        DataFormat = "map-bytes"
	FormatMapString       DataFormat = "map-string"
	FormatMapYamlNodes    DataFormat = "map-yaml-nodes"
	FormatMapNestedString DataFormat = "map-nested-string"
	FormatYaml            DataFormat = "yaml"
	FormatJson            DataFormat = "json"
)

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
