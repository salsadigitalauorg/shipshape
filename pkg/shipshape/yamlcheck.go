package shipshape

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// Merge implementation for Yaml check.
func (c *YamlCheck) Merge(mergeCheck Check) error {
	yCheck := mergeCheck.(*YamlCheck)
	if err := c.YamlBase.Merge(&yCheck.YamlBase); err != nil {
		return err
	}

	utils.MergeString(&c.Path, yCheck.Path)
	utils.MergeString(&c.File, yCheck.File)
	utils.MergeStringSlice(&c.Files, yCheck.Files)
	utils.MergeString(&c.Pattern, yCheck.Pattern)
	utils.MergeString(&c.ExcludePattern, yCheck.ExcludePattern)
	utils.MergeBoolPtrs(c.IgnoreMissing, yCheck.IgnoreMissing)
	return nil
}

// readFile attempts to read a file and assign it to the check's data map using
// the provided file key.
func (c *YamlCheck) readFile(fkey string, fname string) {
	var err error
	c.DataMap[fkey], err = os.ReadFile(fname)
	if err != nil {
		// No failure if missing file and ignoring missing.
		if _, ok := err.(*fs.PathError); ok && c.IgnoreMissing != nil && *c.IgnoreMissing {
			c.AddPass(fmt.Sprintf("File %s does not exist", fname))
			c.Result.Status = Pass
		} else {
			c.AddFail(err.Error())
		}
	}
}

// FetchData populates the DataMap for a File-based Yaml check.
// The check can be run either against a single File, or based on a
// regex Pattern.
func (c *YamlCheck) FetchData() {
	c.DataMap = map[string][]byte{}
	if c.File != "" {
		c.readFile(filepath.Join(c.Path, c.File), filepath.Join(ProjectDir, c.Path, c.File))
	} else if len(c.Files) > 0 {
		for _, f := range c.Files {
			c.readFile(filepath.Join(c.Path, f), filepath.Join(ProjectDir, c.Path, f))
		}
	} else if c.Pattern != "" {
		configPath := filepath.Join(ProjectDir, c.Path)
		files, err := utils.FindFiles(configPath, c.Pattern, c.ExcludePattern, nil)
		if err != nil {
			// No failure if missing path and ignoring missing.
			if _, ok := err.(*fs.PathError); ok && c.IgnoreMissing != nil && *c.IgnoreMissing {
				c.AddPass(fmt.Sprintf("Path %s does not exist", configPath))
				c.Result.Status = Pass
			} else {
				c.AddFail(err.Error())
			}
			return
		}

		if len(files) == 0 && c.IgnoreMissing != nil && *c.IgnoreMissing {
			c.AddPass("no matching config files found")
			c.Result.Status = Pass
			return
		} else if len(files) == 0 {
			c.AddFail("no matching config files found")
			return
		}

		c.DataMap = map[string][]byte{}
		for _, fname := range files {
			c.readFile(fname, fname)
		}
	} else {
		c.AddFail("no file provided")
	}
}
