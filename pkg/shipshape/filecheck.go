package shipshape

import (
	"fmt"
	"path/filepath"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// Merge implementation for file check.
func (c *FileCheck) Merge(mergeCheck Check) error {
	if err := c.CheckBase.Merge(mergeCheck); err != nil {
		return err
	}
	fileMergeCheck := mergeCheck.(*FileCheck)

	utils.MergeString(&c.Path, fileMergeCheck.Path)
	utils.MergeString(&c.DisallowedPattern, fileMergeCheck.DisallowedPattern)
	return nil
}

// RequiresData implementation for file check.
// Since this check acts on the existence of files on disk, it does not require
// any data.
func (c *FileCheck) RequiresData() bool { return false }

// RunCheck implements the check logic for FileCheck.
// It scans a directory for a list of disallowed files and fails it finds any,
// otherwise passes.
func (c *FileCheck) RunCheck() {
	files, err := utils.FindFiles(filepath.Join(ProjectDir, c.Path), c.DisallowedPattern, "")
	if err != nil {
		c.AddFail(err.Error())
		return
	}
	if len(files) == 0 {
		c.Result.Status = Pass
		c.AddPass("No illegal files")
		return
	}
	for _, f := range files {
		c.AddFail(fmt.Sprintf("Illegal file found: %s", f))
	}
}
