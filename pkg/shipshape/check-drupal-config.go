package shipshape

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

func (c *DrupalConfigBase) RunCheck() error {
	if c.Data == nil {
		c.Result.Error = "no data to run check on"
		return nil
	}

	err := c.UnmarshalData(c.Data)
	if err != nil {
		return err
	}

	for _, cv := range c.Values {
		kvr, err := c.CheckKeyValue(cv)
		switch kvr {
		case KeyValueError:
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
			c.Result.Status = Fail
		case KeyValueNotFound:
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("No value found for '%s'", cv.Key),
			)
			c.Result.Status = Fail
		case KeyValueNotEqual:
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("'%s' is not equal to '%s'", cv.Key, cv.Value),
			)
			c.Result.Status = Fail
		case KeyValueEqual:
			c.Result.Passes = append(
				c.Result.Passes,
				fmt.Sprintf("'%s' is equal to '%s'", cv.Key, cv.Value),
			)
			if c.Result.Status == "" {
				c.Result.Status = Pass
			}
		}
	}
	return nil
}

func (c *DrupalFileConfigCheck) FetchData() error {
	var err error
	fullpath := filepath.Join(c.ProjectDir, c.ConfigPath, c.ConfigName+".yml")
	c.Data, err = ioutil.ReadFile(fullpath)
	if err != nil {
		return err
	}
	return nil
}

func (c *DrupalFileModuleCheck) RunCheck() error {
	if c.Data == nil {
		c.Result.Error = "no data to run check on"
		return nil
	}

	err := c.UnmarshalData(c.Data)
	if err != nil {
		return err
	}

	for _, m := range c.Required {
		kvr, err := c.CheckKeyValue(KeyValue{
			Key:   "module." + m,
			Value: "0",
		})
		// It could be a value different from 0, which still means it's enabled.
		if kvr == KeyValueEqual || kvr == KeyValueNotEqual {
			c.Result.Passes = append(
				c.Result.Passes,
				fmt.Sprintf("'%s' is enabled", m),
			)
			if c.Result.Status == "" {
				c.Result.Status = Pass
			}
		} else if kvr == KeyValueError {
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
			c.Result.Status = Fail
		} else {
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("'%s' is not enabled", m),
			)
			c.Result.Status = Fail
		}
	}
	for _, m := range c.Disallowed {
		kvr, err := c.CheckKeyValue(KeyValue{
			Key:   "module." + m,
			Value: "0",
		})
		// It could be a value different from 0, which still means it's enabled.
		if kvr == KeyValueEqual || kvr == KeyValueNotEqual {
			c.Result.Failures = append(
				c.Result.Failures,
				fmt.Sprintf("'%s' is enabled", m),
			)
			c.Result.Status = Fail
		} else if kvr == KeyValueError {
			c.Result.Failures = append(
				c.Result.Failures,
				err.Error(),
			)
			c.Result.Status = Fail
		} else {
			c.Result.Passes = append(
				c.Result.Passes,
				fmt.Sprintf("'%s' is not enabled", m),
			)
			if c.Result.Status == "" {
				c.Result.Status = Pass
			}
		}
	}
	err = c.DrupalFileConfigCheck.RunCheck()
	return err
}

func (c *DrupalFileModuleCheck) Init(pd string, ct CheckType) {
	c.CheckBase.Init(pd, ct)
	c.ConfigName = "core.extension"
}
