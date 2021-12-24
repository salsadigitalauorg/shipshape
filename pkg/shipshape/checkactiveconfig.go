package shipshape

import "fmt"

func (c *ActiveConfigCheck) RunCheck() error {
	fmt.Println("Running active config check")
	return nil
}

func (c *ActiveConfigCheck) GetResult() Result {
	r := Result{}
	return r
}
