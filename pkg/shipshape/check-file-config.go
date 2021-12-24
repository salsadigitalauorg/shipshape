package shipshape

import "fmt"

func (c *FileConfigCheck) FetchData() error {
	c.Data = []byte("")
	return nil
}

func (c *FileConfigCheck) RunCheck() (Result, error) {
	r := Result{CheckType: FileConfig}
	fmt.Println("Running file config check")
	return r, nil
}
