package file

import (
	"errors"
	"fmt"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
	"io/fs"
	"os"
	"path/filepath"
)

type FileDiffCheck struct {
	config.CheckBase `yaml:",inline"`
	// TargetFile will be compared with SourceFile.
	TargetFile string `yaml:"target"`
	// SourceFile can be a local file or a remote URI.
	SourceFile string `yaml:"source"`
	// SourceContext list of key-values to compile the source file as a Jinja template.
	SourceContext map[string]any `yaml:"source-context"`
	// ContextLines number of context lines around the line changes.
	ContextLines int `yaml:"context-lines"`
	// IgnoreMissing allows non-existent files to not be counted as a Fail.
	// Using a pointer here so that we can differentiate between
	// false (default value) and an empty value.
	IgnoreMissing *bool `yaml:"ignore-missing"`
}

const FileDiff config.CheckType = "filediff"

// RequiresData implementation for FileDiffCheck.
func (c *FileDiffCheck) RequiresData() bool { return true }

// FetchData implementation for FileDiffCheck
func (c *FileDiffCheck) FetchData() {
	if len(c.SourceFile) == 0 {
		c.AddBreach(&result.ValueBreach{Value: "no source file provided"})
		return
	}

	if len(c.TargetFile) == 0 {
		c.AddBreach(&result.ValueBreach{Value: "no target file provided"})
		return
	}

	c.DataMap = map[string][]byte{}
	var err error
	// Fetch the target file.
	c.DataMap["target"], err = os.ReadFile(filepath.Join(config.ProjectDir, c.TargetFile))
	if err != nil {
		// No failure if missing file and ignoring missing.
		var pathError *fs.PathError
		if errors.As(err, &pathError) && c.IgnoreMissing != nil && *c.IgnoreMissing {
			c.AddPass(fmt.Sprintf("Target file %s does not exist", c.TargetFile))
			c.Result.Status = result.Pass
			return
		} else {
			c.AddBreach(&result.ValueBreach{
				ValueLabel: "error reading target file: " + c.TargetFile,
				Value:      err.Error()})
			return
		}
	}

	// Fetch the source file.
	if utils.StringIsUrl(c.SourceFile) {
		c.DataMap["source"], err = utils.FetchContentFromUrl(c.SourceFile)
	} else {
		c.DataMap["source"], err = os.ReadFile(filepath.Join(config.ProjectDir, c.SourceFile))
	}

	if err != nil {
		c.AddBreach(&result.ValueBreach{
			ValueLabel: "error fetching source file: " + c.SourceFile,
			Value:      err.Error()})
		return
	}

	// Parse the source file as a Jinja template.
	if c.SourceContext != nil && len(c.SourceContext) > 0 {
		jinjaTemplate, jinjaErr := gonja.FromBytes(c.DataMap["source"])
		if jinjaErr != nil {
			c.AddBreach(&result.ValueBreach{
				ValueLabel: "error parsing source file: " + c.SourceFile,
				Value:      jinjaErr.Error()})
			return
		}

		jinjaContext := exec.NewContext(c.SourceContext)
		c.DataMap["source"], jinjaErr = jinjaTemplate.ExecuteToBytes(jinjaContext)
		if jinjaErr != nil {
			c.AddBreach(&result.ValueBreach{
				ValueLabel: "error compiling source file with source context: " + c.SourceFile,
				Value:      jinjaErr.Error()})
			return
		}
	}

	return
}

// Merge implementation for FileDiffCheck check.
func (c *FileDiffCheck) Merge(mergeCheck config.Check) error {
	yCheck := mergeCheck.(*FileDiffCheck)
	if err := c.CheckBase.Merge(&yCheck.CheckBase); err != nil {
		return err
	}

	if yCheck.ContextLines != 0 && yCheck.ContextLines != c.ContextLines {
		c.ContextLines = yCheck.ContextLines
	}

	if yCheck.SourceContext != nil && len(yCheck.SourceContext) != 0 {
		c.SourceContext = yCheck.SourceContext
	}

	utils.MergeString(&c.SourceFile, yCheck.SourceFile)
	utils.MergeString(&c.TargetFile, yCheck.TargetFile)
	utils.MergeBoolPtrs(c.IgnoreMissing, yCheck.IgnoreMissing)
	return nil
}

// UnmarshalDataMap implementation for FileDiffCheck check.
func (c *FileDiffCheck) UnmarshalDataMap() {}

// RunCheck implementation for FileDiffCheck check.
func (c *FileDiffCheck) RunCheck() {
	unifiedDiff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(c.DataMap["source"])),
		B:        difflib.SplitLines(string(c.DataMap["target"])),
		FromFile: c.SourceFile,
		ToFile:   c.TargetFile,
		Context:  c.ContextLines,
	}
	diff, _ := difflib.GetUnifiedDiffString(unifiedDiff)
	if len(diff) == 0 {
		c.AddPass(fmt.Sprintf("Target file %s is identical to Source file %s", c.TargetFile, c.SourceFile))
		c.Result.Status = result.Pass
	} else {
		c.AddBreach(&result.ValueBreach{
			ValueLabel: fmt.Sprintf("Target file %s is different from Source file %s", c.TargetFile, c.SourceFile),
			Value:      fmt.Sprintf("diff: \n%s", diff)})
	}
}
