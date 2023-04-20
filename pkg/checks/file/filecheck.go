package file

import (
	"fmt"
	"path/filepath"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

//go:generate go run ../../../cmd/gen.go registry --checkpackage=file

// FileCheck is a simple File absence check which can be for a single
// file or a pattern.
type FileCheck struct {
	config.CheckBase  `yaml:",inline"`
	Path              string   `yaml:"path"`
	DisallowedPattern string   `yaml:"disallowed-pattern"`
	ExcludePattern    string   `yaml:"exclude-pattern"`
	SkipDir           []string `yaml:"skip-dir"`
}

const File config.CheckType = "file"

func RegisterChecks() {
	config.ChecksRegistry[File] = func() config.Check { return &FileCheck{} }
}

func init() {
	RegisterChecks()
}

// Merge implementation for file check.
func (c *FileCheck) Merge(mergeCheck config.Check) error {
	fileMergeCheck := mergeCheck.(*FileCheck)
	if err := c.CheckBase.Merge(&fileMergeCheck.CheckBase); err != nil {
		return err
	}

	utils.MergeString(&c.Path, fileMergeCheck.Path)
	utils.MergeString(&c.DisallowedPattern, fileMergeCheck.DisallowedPattern)
	return nil
}

// RequiresData implementation for file check.
// Since this check acts on the existence of files on disk, it does not require
// any data.
func (c *FileCheck) RequiresData() bool { return false }

// RunCheck scans a directory for a list of disallowed files, while excluding
// the provided regex ExcludePattern and skipping the list of provided relative
// directories.
func (c *FileCheck) RunCheck() {
	files, err := utils.FindFiles(filepath.Join(config.ProjectDir, c.Path), c.DisallowedPattern, c.ExcludePattern, c.SkipDir)
	if err != nil {
		c.AddFail(err.Error())
		return
	}
	if len(files) == 0 {
		c.Result.Status = result.Pass
		c.AddPass("No illegal files")
		return
	}
	for _, f := range files {
		c.AddFail(fmt.Sprintf("Illegal file found: %s", f))
	}
}
