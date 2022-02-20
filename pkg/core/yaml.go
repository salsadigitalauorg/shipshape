package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"salsadigitalauorg/shipshape/pkg/utils"
	"strings"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

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

func (y *YamlBase) RunCheck() {
	if y.DataMap == nil {
		y.Result.Status = Fail
		y.Result.Failures = append(
			y.Result.Failures,
			"no data available",
		)
		return
	}

	err := y.UnmarshalDataMap()
	if err != nil {
		y.Result.Status = Fail
		y.Result.Failures = append(
			y.Result.Failures,
			err.Error(),
		)
		return
	}

	for configName := range y.DataMap {
		y.processData(configName)
	}
}

func (y *YamlBase) UnmarshalDataMap() error {
	y.NodeMap = map[string]yaml.Node{}
	for configName, data := range y.DataMap {
		n := yaml.Node{}
		err := yaml.Unmarshal([]byte(data), &n)
		if err != nil {
			return err
		}
		y.NodeMap[configName] = n
	}
	return nil
}

func (y *YamlBase) processData(configName string) {
	for _, kv := range y.Values {
		kvr, fails, err := y.CheckKeyValue(kv, configName)
		switch kvr {
		case KeyValueError:
			y.Result.Failures = append(
				y.Result.Failures,
				err.Error(),
			)
		case KeyValueNotFound:
			y.Result.Failures = append(
				y.Result.Failures,
				fmt.Sprintf("[%s] '%s' not found", configName, kv.Key),
			)
		case KeyValueNotEqual:
			y.Result.Failures = append(
				y.Result.Failures,
				fmt.Sprintf("[%s] '%s' equals '%s'", configName, kv.Key, fails[0]),
			)
		case KeyValueDisallowedFound:
			y.Result.Failures = append(
				y.Result.Failures,
				fmt.Sprintf("[%s] disallowed '%s': [%s]", configName, kv.Key, strings.Join(fails, ", ")),
			)
		case KeyValueEqual:
			if kv.IsList {
				y.Result.Passes = append(
					y.Result.Passes,
					fmt.Sprintf("[%s] no disallowed '%s'", configName, kv.Key),
				)
			} else {
				y.Result.Passes = append(
					y.Result.Passes,
					fmt.Sprintf("[%s] '%s' equals '%s'", configName, kv.Key, kv.Value),
				)
			}
		}
	}
	if len(y.Result.Failures) != 0 {
		y.Result.Status = Fail
	} else {
		y.Result.Status = Pass
	}
}

func (y *YamlBase) CheckKeyValue(kv KeyValue, mapKey string) (KeyValueResult, []string, error) {
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

func (y *YamlCheck) FetchData() {
	var err error
	y.DataMap = map[string][]byte{}
	if y.File != "" {
		fullPath := filepath.Join(y.ProjectDir, y.Path, y.File+".yml")
		y.DataMap[y.File+".yml"], err = ioutil.ReadFile(fullPath)
		if err != nil {
			y.Result.Status = Fail
			y.Result.Failures = append(
				y.Result.Failures,
				err.Error(),
			)
		}
	} else if y.Pattern != "" {
		configPath := filepath.Join(y.ProjectDir, y.Path)
		files, err := utils.FindFiles(configPath, y.Pattern, y.ExcludePattern)
		if err != nil {
			y.Result.Status = Fail
			y.Result.Failures = append(
				y.Result.Failures,
				err.Error(),
			)
			return
		}

		if len(files) == 0 {
			y.Result.Status = Fail
			y.Result.Failures = append(
				y.Result.Failures,
				"no matching config files found",
			)
			return
		}

		y.DataMap = map[string][]byte{}
		for _, fname := range files {
			l, err := ioutil.ReadFile(fname)
			if err != nil {
				y.Result.Status = Fail
				y.Result.Failures = append(
					y.Result.Failures,
					err.Error(),
				)
			}
			_, file := filepath.Split(fname)
			y.DataMap[file] = l
		}
	} else {
		y.Result.Status = Fail
		y.Result.Failures = append(
			y.Result.Failures,
			"no config file name provided",
		)
	}
}
