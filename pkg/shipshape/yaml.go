package shipshape

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"

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
			c.AddFail(fmt.Sprintf("[%s] disallowed %s: [%s]", configName, kv.Key, strings.Join(fails, ", ")))
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

// CheckAllowDisallowList validates against allow/disallow lists and returns
// true if a disallowed value is present.
func CheckAllowDisallowList(kv KeyValue, value string) (bool) {

	// Ignore blank and null values.
	if(len(value) == 0) {
		return false
	}

	// Check disallowed list.
	if len(kv.Disallowed) > 0 && utils.StringSliceContains(kv.Disallowed, value) {
		return true
	}

	// Check allowed list.
	if len(kv.Allowed) > 0 && !utils.StringSliceContains(kv.Allowed, value) {
		return true
	}

	return false
}

// CheckKeyValue lookups the Yaml data for a specific KeyValue and returns the
// result, actual values and errors.
func (c *YamlBase) CheckKeyValue(kv KeyValue, mapKey string) (KeyValueResult, []string, error) {
	node := c.NodeMap[mapKey]
	foundNodes, err := LookupYamlPath(&node, kv.Key)
	if err != nil {
		return KeyValueError, nil, err
	}

	if len(foundNodes) == 0 && !kv.Optional {
		return KeyValueNotFound, nil, nil
	}

	// Throw an error if we are checking a list but no allow/disallow list provided.
	if len(kv.Allowed) == 0 && len(kv.Disallowed) == 0 && kv.IsList {
		return KeyValueError, nil, errors.New("list of allowed or disallowed values not provided")
	}

	// Perform direct comparison if no allow/disallow list provided.
	if len(kv.Allowed) == 0 && len(kv.Disallowed) == 0 {
		for _, item := range foundNodes {
			notEquals := []string{}
			// When checking for false, "null" is also 'falsy'.
			if item.Value != kv.Value && (kv.Value != "false" || item.Value != "null") {
				if !utils.StringSliceContains(notEquals, item.Value) {
					notEquals = append(notEquals, item.Value)
				}
			}
			if len(notEquals) > 0 {
				return KeyValueNotEqual, notEquals, nil
			}
		}
		return KeyValueEqual, nil, nil
	}

	// Check each yaml value against the disallowed list.
	fails := []string{}
	for _, item := range foundNodes {
		if kv.IsList {
			for _, v := range item.Content {
				if CheckAllowDisallowList(kv, v.Value) && !utils.StringSliceContains(fails, v.Value) {
					fails = append(fails, v.Value)
				}
			}
		} else {
			if CheckAllowDisallowList(kv, item.Value) && !utils.StringSliceContains(fails, item.Value) {
				fails = append(fails, item.Value)
			}
		}
	}
	if len(fails) > 0 {
		return KeyValueDisallowedFound, fails, nil
	}
	return KeyValueEqual, nil, nil
}

// readFile attempts to read a file and assign it to the check's data map using
// the provided file key.
func (c *YamlCheck) readFile(fkey string, fname string) {
	var err error
	c.DataMap[fkey], err = ioutil.ReadFile(fname)
	if err != nil {
		// No failure if missing file and ignoring missing.
		if _, ok := err.(*fs.PathError); ok && c.IgnoreMissing {
			c.AddPass(fmt.Sprintf("File %s does not exist", fname))
			c.Result.Status = Pass
		} else {
			c.AddFail(err.Error())
		}
	}
}

// FetchData populates the DataMap for a File-based Yaml check.
// The check can be run either against a single File, or based on a
// regex Pattern.
func (c *YamlCheck) FetchData() {
	c.DataMap = map[string][]byte{}
	if c.File != "" {
		c.readFile(filepath.Join(c.Path, c.File), filepath.Join(ProjectDir, c.Path, c.File))
	} else if len(c.Files) > 0 {
		for _, f := range c.Files {
			c.readFile(filepath.Join(c.Path, f), filepath.Join(ProjectDir, c.Path, f))
		}
	} else if c.Pattern != "" {
		configPath := filepath.Join(ProjectDir, c.Path)
		files, err := utils.FindFiles(configPath, c.Pattern, c.ExcludePattern)
		if err != nil {
			// No failure if missing path and ignoring missing.
			if _, ok := err.(*fs.PathError); ok && c.IgnoreMissing {
				c.AddPass(fmt.Sprintf("Path %s does not exist", configPath))
				c.Result.Status = Pass
			} else {
				c.AddFail(err.Error())
			}
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
			_, file := filepath.Split(fname)
			c.readFile(filepath.Join(c.Path, file), fname)
		}
	} else {
		c.AddFail("no file provided")
	}
}

// UnmarshalDataMap tries to parse the yaml file into a generic structure and
// returns any errors as failures.
func (c *YamlLintCheck) UnmarshalDataMap() {
	for f, data := range c.DataMap {
		ifc := make(map[string]interface{})
		err := yaml.Unmarshal([]byte(data), &ifc)
		if err != nil {
			if typeErr, ok := err.(*yaml.TypeError); ok {
				for _, msg := range typeErr.Errors {
					c.AddFail(fmt.Sprintf("[%s] %s", f, msg))
				}
			} else {
				c.AddFail(fmt.Sprintf("[%s] %s", f, err.Error()))
			}
		} else {
			c.AddPass(fmt.Sprintf("%s has valid yaml.", f))
		}
	}
	if c.Result.Status != Fail {
		c.Result.Status = Pass
	}
}

// RunCheck for YamlLint does nothing since the check is in UnmarshalDataMap.
func (c *YamlLintCheck) RunCheck() {}
