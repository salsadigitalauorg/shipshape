package core

import (
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func (y *YamlCheck) UnmarshalData(data []byte) error {
	err := yaml.Unmarshal([]byte(data), &y.Node)
	if err != nil {
		return err
	}
	return nil
}

func (y *YamlCheck) CheckKeyValue(kv KeyValue) (KeyValueResult, error) {
	q, err := LookupYamlPath(&y.Node, kv.Key)
	if err != nil {
		return KeyValueError, err
	}

	if len(q) == 0 {
		return KeyValueNotFound, nil
	}

	if q[0].Value != kv.Value {
		return KeyValueNotEqual, nil
	}

	return KeyValueEqual, nil
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
