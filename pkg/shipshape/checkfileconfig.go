package shipshape

import "fmt"

func (c *FileConfigCheck) RunCheck() error {
	fmt.Println("Running file config check")
	return nil
}

func (c *FileConfigCheck) GetResult() Result {
	r := Result{}
	return r
}
