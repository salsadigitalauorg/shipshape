package drupal

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"salsadigitalauorg/shipshape/pkg/core"
	"salsadigitalauorg/shipshape/pkg/utils"
	"strings"
)

func (c *ConfigBase) RunCheck() {
	if c.DataMap == nil {
		c.Result.Status = core.Fail
		c.Result.Failures = append(
			c.Result.Failures,
			"no data to run check on",
		)
		return
	}

	err := c.UnmarshalDataMap(c.DataMap)
	if err != nil {
		c.Result.Status = core.Fail
		c.Result.Failures = append(
			c.Result.Failures,
			err.Error(),
		)
		return
	}
	c.processDataMap()
}

func (c *ConfigBase) processData(configName string) {
	for _, kv := range c.Values {
		kvr, fails, err := c.CheckKeyValue(kv, configName)
		switch kvr {
		case core.KeyValueError:
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
		case core.KeyValueNotFound:
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("[%s] '%s' not found", configName, kv.Key),
			)
		case core.KeyValueNotEqual:
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("[%s] '%s' equals '%s'", configName, kv.Key, fails[0]),
			)
		case core.KeyValueDisallowedFound:
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("[%s] disallowed '%s': [%s]", configName, kv.Key, strings.Join(fails, ", ")),
			)
		case core.KeyValueEqual:
			if kv.IsList {
				c.Result.Passes = append(
					c.Result.Passes,
					fmt.Sprintf("[%s] no disallowed '%s'", configName, kv.Key),
				)
				continue
			}
			c.Result.Passes = append(
				c.Result.Passes,
				fmt.Sprintf("[%s] '%s' equals '%s'", configName, kv.Key, kv.Value),
			)
			if c.Result.Status == "" {
				c.Result.Status = core.Pass
			}
		}
	}
	if c.Result.Status == "" {
		c.Result.Status = core.Fail
	}
}

func (c *ConfigBase) processDataMap() {
	for configName := range c.DataMap {
		c.processData(configName)
	}
}

func (c *FileConfigCheck) FetchData() {
	var err error
	c.DataMap = map[string][]byte{}
	if c.ConfigName != "" {
		fullPath := filepath.Join(c.ProjectDir, c.Path, c.ConfigName+".yml")
		c.DataMap[c.ConfigName+".yml"], err = ioutil.ReadFile(fullPath)
		if err != nil {
			c.Result.Status = core.Fail
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
		}
	} else if c.Pattern != "" {
		configPath := filepath.Join(c.ProjectDir, c.Path)
		files, err := utils.FindFiles(configPath, c.Pattern, c.ExcludePattern)
		if err != nil {
			c.Result.Status = core.Fail
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
		}

		if len(files) == 0 {
			c.Result.Status = core.Fail
			c.Result.Failures = append(
				c.Result.Failures,
				"no matching config files found",
			)
		}

		c.DataMap = map[string][]byte{}
		for _, fname := range files {
			l, err := ioutil.ReadFile(fname)
			if err != nil {
				c.Result.Status = core.Fail
				c.Result.Failures = append(
					c.Result.Failures,
					err.Error(),
				)
			}
			_, file := filepath.Split(fname)
			c.DataMap[file] = l
		}
	} else {
		c.Result.Status = core.Fail
		c.Result.Failures = append(
			c.Result.Failures,
			"no config file name provided",
		)
	}
}

func (c *FileModuleCheck) RunCheck() {
	if c.DataMap == nil {
		c.Result.Status = core.Fail
		c.Result.Failures = append(
			c.Result.Failures,
			"no data to run check on",
		)
		return
	}

	err := c.UnmarshalDataMap(c.DataMap)
	if err != nil {
		c.Result.Status = core.Fail
		c.Result.Failures = append(
			c.Result.Failures,
			err.Error(),
		)
		return
	}

	for _, m := range c.Required {
		kvr, _, err := c.CheckKeyValue(core.KeyValue{
			Key:   "module." + m,
			Value: "0",
		}, c.ConfigName+".yml")
		// It could be a value different from 0, which still means it's enabled.
		if kvr == core.KeyValueEqual || kvr == core.KeyValueNotEqual {
			c.Result.Passes = append(
				c.Result.Passes,
				fmt.Sprintf("'%s' is enabled", m),
			)
			if c.Result.Status == "" {
				c.Result.Status = core.Pass
			}
		} else if kvr == core.KeyValueError {
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
			c.Result.Status = core.Fail
		} else {
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("'%s' is not enabled", m),
			)
			c.Result.Status = core.Fail
		}
	}
	for _, m := range c.Disallowed {
		kvr, _, err := c.CheckKeyValue(core.KeyValue{
			Key:   "module." + m,
			Value: "0",
		}, c.ConfigName+".yml")
		// It could be a value different from 0, which still means it's enabled.
		if kvr == core.KeyValueEqual || kvr == core.KeyValueNotEqual {
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("'%s' is enabled", m),
			)
			c.Result.Status = core.Fail
		} else if kvr == core.KeyValueError {
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
			c.Result.Status = core.Fail
		} else {
			c.Result.Passes = append(
				c.Result.Passes,
				fmt.Sprintf("'%s' is not enabled", m),
			)
			if c.Result.Status == "" {
				c.Result.Status = core.Pass
			}
		}
	}
	c.FileConfigCheck.RunCheck()
}

func (c *FileModuleCheck) Init(pd string, ct core.CheckType) {
	c.CheckBase.Init(pd, ct)
	c.ConfigName = "core.extension"
}

func (c *ActiveModuleCheck) RunCheck() {
	if c.DataMap == nil {
		c.Result.Status = core.Fail
		c.Result.Failures = append(
			c.Result.Failures,
			"no data to run check on",
		)
		return
	}
}
