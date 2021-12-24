package shipshape

import "fmt"

func (c *DrushCommandCheck) FetchData() error {
	c.Data = []byte("")
	return nil
}

func (c *DrushCommandCheck) RunCheck() (Result, error) {
	r := Result{CheckType: DrushCommand}
	fmt.Println("Running drush command check")
	return r, nil
}
