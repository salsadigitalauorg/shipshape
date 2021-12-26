package shipshape

import (
	"errors"
	"fmt"
)

func (c *DrupalConfigBase) FetchData() error {
	c.Data = []byte("")
	return nil
}

func (c *DrupalConfigBase) RunCheck() error {
	c.Result = Result{CheckType: DrupalDBConfig}

	if c.Data == nil {
		return errors.New("no data to run check on")
	}

	c.UnmarshalData(c.Data)

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
