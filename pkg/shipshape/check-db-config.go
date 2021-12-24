package shipshape

import (
	"errors"
	"fmt"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func (c *DbConfigCheck) FetchData() error {
	c.Data = []byte("")
	return nil
}

func (c *DbConfigCheck) RunCheck() error {
	c.Result = Result{CheckType: DbConfig}

	if c.Data == nil {
		return errors.New("no data to run check on")
	}

	var n yaml.Node
	err := yaml.Unmarshal([]byte(c.Data), &n)
	if err != nil {
		return nil
	}

	for _, cv := range c.ConfigValues {
		c.CheckKeyValue(&n, cv.Key, cv.Value)
	}
	return nil
}

func (c *DbConfigCheck) CheckKeyValue(n *yaml.Node, key string, value string) {
	p, err := yamlpath.NewPath(key)
	if err != nil {
		c.Result.Failures = append(
			c.Result.Failures,
			"unable to construct yaml path",
		)
		c.Result.Status = Fail
		return
	}

	q, err := p.Find(n)
	if err != nil {
		c.Result.Failures = append(
			c.Result.Failures,
			"encountered error while looking up yaml path",
		)
		c.Result.Status = Fail
		return
	}

	if len(q) == 0 {
		c.Result.Failures = append(
			c.Result.Failures,
			fmt.Sprintf("No value found for '%s'", key),
		)
		c.Result.Status = Fail
		return
	}

	if q[0].Value != value {
		c.Result.Failures = append(
			c.Result.Failures,
			fmt.Sprintf("'%s' is not equal to '%s'", key, value),
		)
		c.Result.Status = Fail
		return
	}

	c.Result.Passes = append(
		c.Result.Passes,
		fmt.Sprintf("'%s' is equal to '%s'", key, value),
	)

	if c.Result.Status == "" {
		c.Result.Status = Pass
	}
}
