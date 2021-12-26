package shipshape

import "fmt"

func (c *DrupalFileConfigCheck) FetchData() error {
	c.Data = []byte("")
	return nil
}

func (c *DrupalFileConfigCheck) RunCheck() (Result, error) {
	r := Result{CheckType: DrupalFileConfig}
	fmt.Println("Running file config check")
	return r, nil
}
