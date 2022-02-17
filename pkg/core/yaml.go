package core

import (
	"errors"
	"salsadigitalauorg/shipshape/pkg/utils"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func (y *YamlCheck) UnmarshalDataMap(dataMap map[string][]byte) error {
	y.NodeMap = map[string]yaml.Node{}
	for configName, data := range dataMap {
		n := yaml.Node{}
		err := yaml.Unmarshal([]byte(data), &n)
		if err != nil {
			return err
		}
		y.NodeMap[configName] = n
	}
	return nil
}

func (y *YamlCheck) CheckKeyValue(kv KeyValue, mapKey string) (KeyValueResult, []string, error) {
	node := y.NodeMap[mapKey]
	q, err := LookupYamlPath(&node, kv.Key)
	if err != nil {
		return KeyValueError, nil, err
	}

	if len(q) == 0 {
		return KeyValueNotFound, nil, nil
	}

	if !kv.IsList {
		// When checking for false, "null" is also 'falsy'.
		if q[0].Value != kv.Value && (kv.Value != "false" || q[0].Value != "null") {
			return KeyValueNotEqual, []string{q[0].Value}, nil
		}
		return KeyValueEqual, nil, nil
	}

	if len(kv.Disallowed) == 0 {
		return KeyValueError, nil, errors.New("list of disallowed values not provided")
	}

	// Check each yaml value against the disallowed list.
	fails := []string{}
	for _, v := range q[0].Content {
		if utils.StringSliceContains(kv.Disallowed, v.Value) {
			fails = append(fails, v.Value)
		}
	}
	if len(fails) > 0 {
		return KeyValueDisallowedFound, fails, nil
	}
	return KeyValueEqual, nil, nil
}

func LookupYamlPath(y *yaml.Node, path string) ([]*yaml.Node, error) {
	p, err := yamlpath.NewPath(path)
	if err != nil {
		return nil, err
	}
	q, err := p.Find(y)
	if err != nil {
		return nil, err
	}
	return q, nil
}
