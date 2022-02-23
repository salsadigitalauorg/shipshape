package core

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"salsadigitalauorg/shipshape/pkg/utils"
	"strings"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

// LookupYamlPath attempts to query Yaml data using a JSONPath query and returns
// the found Node.
// It uses the implemention by https://github.com/vmware-labs/yaml-jsonpath.
func LookupYamlPath(n *yaml.Node, path string) ([]*yaml.Node, error) {
	p, err := yamlpath.NewPath(path)
	if err != nil {
		return nil, err
	}
	q, err := p.Find(n)
	if err != nil {
		return nil, err
	}
	return q, nil
}

// RunCheck implements the base logic for running checks against Yaml data.
func (c *YamlBase) RunCheck() {
	for configName := range c.DataMap {
		c.processData(configName)
	}
}

// UnmarshalDataMap parses the DataMap into Yaml for further processing.
// DataMap is expected to be populated from FetchData in the respective Check
// implementation.
func (c *YamlBase) UnmarshalDataMap() {
	c.NodeMap = map[string]yaml.Node{}
	for configName, data := range c.DataMap {
		n := yaml.Node{}
		err := yaml.Unmarshal([]byte(data), &n)
		if err != nil {
			c.AddFail(err.Error())
			return
		}
		c.NodeMap[configName] = n
	}
}

// processData runs the actual checks against the list of KeyValues provided in
// the Check configuration and determines the Status (Pass or Fail) and Pass or
// Fail messages of the Check Result.
func (c *YamlBase) processData(configName string) {
	for _, kv := range c.Values {
		kvr, fails, err := c.CheckKeyValue(kv, configName)
		switch kvr {
		case KeyValueError:
			c.AddFail(err.Error())
		case KeyValueNotFound:
			c.AddFail(fmt.Sprintf("[%s] '%s' not found", configName, kv.Key))
		case KeyValueNotEqual:
			c.AddFail(fmt.Sprintf("[%s] '%s' equals '%s'", configName, kv.Key, fails[0]))
		case KeyValueDisallowedFound:
			c.AddFail(fmt.Sprintf("[%s] disallowed '%s': [%s]", configName, kv.Key, strings.Join(fails, ", ")))
		case KeyValueEqual:
			if kv.IsList {
				c.AddPass(fmt.Sprintf("[%s] no disallowed '%s'", configName, kv.Key))
			} else {
				c.AddPass(fmt.Sprintf("[%s] '%s' equals '%s'", configName, kv.Key, kv.Value))
			}
		}
	}
	if len(c.Result.Failures) != 0 {
		c.Result.Status = Fail
	} else {
		c.Result.Status = Pass
	}
}

// CheckKeyValue lookups the Yaml data for a specific KeyValue and returns the
// result, actual values and errors.
func (c *YamlBase) CheckKeyValue(kv KeyValue, mapKey string) (KeyValueResult, []string, error) {
	node := c.NodeMap[mapKey]
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

// FetchData populates the DataMap for a File-based Yaml check.
// The check can be run either against a single File, or based on a
// regex Pattern.
func (c *YamlCheck) FetchData() {
	var err error
	c.DataMap = map[string][]byte{}
	if c.File != "" {
		fullPath := filepath.Join(ProjectDir, c.Path, c.File)
		c.DataMap[c.File], err = ioutil.ReadFile(fullPath)
		if err != nil {
			// No failure if missing file and ignoring missing.
			if _, ok := err.(*fs.PathError); ok && c.IgnoreMissing {
				c.AddPass("File does not exist")
				c.Result.Status = Pass
			} else {
				c.AddFail(err.Error())
			}
		}
	} else if c.Pattern != "" {
		configPath := filepath.Join(ProjectDir, c.Path)
		files, err := utils.FindFiles(configPath, c.Pattern, c.ExcludePattern)
		if err != nil {
			c.AddFail(err.Error())
			return
		}

		if len(files) == 0 && c.IgnoreMissing {
			c.AddPass("no matching config files found")
			c.Result.Status = Pass
			return
		} else if len(files) == 0 {
			c.AddFail("no matching config files found")
			return
		}

		c.DataMap = map[string][]byte{}
		for _, fname := range files {
			l, err := ioutil.ReadFile(fname)
			if err != nil {
				c.AddFail(err.Error())
			}
			_, file := filepath.Split(fname)
			c.DataMap[file] = l
		}
	} else {
		c.AddFail("no config file name provided")
	}
}
