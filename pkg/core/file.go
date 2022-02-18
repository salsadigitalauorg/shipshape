package core

import (
	"fmt"
	"path/filepath"
	"salsadigitalauorg/shipshape/pkg/utils"
)

// RunCheck implements the check logic for FileCheck.
// It scans a directory for a list of disallowed files and fails it finds any,
// otherwise passes.
func (c *FileCheck) RunCheck() {
	files, err := utils.FindFiles(filepath.Join(c.ProjectDir, c.Path), c.DisallowedPattern, "")
	if err != nil {
		c.Result.Status = Fail
		c.Result.Failures = append(
			c.Result.Failures,
			err.Error(),
		)
		return
	}
	if len(files) == 0 {
		c.Result.Status = Pass
		c.Result.Passes = append(
			c.Result.Passes,
			"No illegal files",
		)
		return
	}
	c.Result.Status = Fail
	for _, f := range files {
		c.Result.Failures = append(
			c.Result.Failures,
			fmt.Sprintf("Illegal file found: %s", f),
		)
	}
}
