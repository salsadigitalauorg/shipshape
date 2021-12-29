package shipshape

import (
	"fmt"
	"io/ioutil"
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
	c.Data, err = ioutil.ReadFile(c.ConfigPath)
	if err != nil {
		return err
	}
	return nil
}

func (c *DrupalFileConfigCheck) RunCheck() error {
	c.InitResult(DrupalFileConfig)
	err := c.DrupalConfigBase.RunCheck()
	return err
}

func (c *DrupalDBConfigCheck) FetchData() error {
	return nil
}

func (c *DrupalDBConfigCheck) RunCheck() error {
	c.InitResult(DrupalDBConfig)
	err := c.DrupalConfigBase.RunCheck()
	return err
}

func (c *DrupalFileModuleCheck) RunCheck() error {
	c.InitResult(DrupalModules)
	err := c.DrupalFileConfigCheck.RunCheck()
	return err
}

func (c *DrupalActiveModuleCheck) RunCheck() error {
	c.InitResult(DrupalActiveModules)
	return nil
}
