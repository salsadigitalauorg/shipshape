// Package phpstan provides the types and functions for running a phpstan check
// and parsing its output to process errors.
package phpstan

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

//go:generate go run ../../../cmd/gen.go registry --checkpackage=phpstan

const PhpStan config.CheckType = "phpstan"

type PhpStanCheck struct {
	config.CheckBase `yaml:",inline"`
	Bin              string   `yaml:"binary"`
	Config           string   `yaml:"configuration"`
	Paths            []string `yaml:"paths"`
	phpstanResult    PhpStanResult
}

type PhpStanResult struct {
	Totals struct {
		Errors     int `json:"errors"`
		FileErrors int `json:"file_errors"`
	} `json:"totals"`

	// Keep the raw json, as it could be an empty slice if no errors,
	// a map otherwise.
	FilesRaw json.RawMessage `json:"files"`
	Files    map[string]struct {
		Errors   int `json:"errors"`
		Messages []struct {
			Message   string `json:"message"`
			Line      int    `json:"line"`
			Ignorable bool   `json:"ignorable"`
		} `json:"messages"`
	}
	Errors []string `json:"errors"`
}

func RegisterChecks() {
	config.ChecksRegistry[PhpStan] = func() config.Check { return &PhpStanCheck{} }
}

func init() {
	RegisterChecks()
}

const PhpstanDefaultPath = "vendor/phpstan/phpstan/phpstan"

// Merge implementation for file check.
func (c *PhpStanCheck) Merge(mergeCheck config.Check) error {
	phpstanMergeCheck := mergeCheck.(*PhpStanCheck)
	if err := c.CheckBase.Merge(&phpstanMergeCheck.CheckBase); err != nil {
		return err
	}

	utils.MergeString(&c.Bin, phpstanMergeCheck.Bin)
	utils.MergeString(&c.Config, phpstanMergeCheck.Config)
	utils.MergeStringSlice(&c.Paths, phpstanMergeCheck.Paths)
	return nil
}

func (c *PhpStanCheck) GetBinary() (path string) {
	if len(c.Bin) == 0 {
		path = filepath.Join(config.ProjectDir, PhpstanDefaultPath)
	} else {
		path = c.Bin
	}
	return
}

// FetchData runs the phpstan command to populate data for the check.
func (c *PhpStanCheck) FetchData() {
	var err error
	phpstanPath := c.GetBinary()

	configPath := c.Config
	if !filepath.IsAbs(c.Config) {
		configPath = filepath.Join(config.ProjectDir, configPath)
	}

	args := []string{
		"analyse",
		"--configuration=" + configPath,
		"--no-progress",
		"--error-format=json",
	}
	foundPath := false
	for _, p := range c.Paths {
		path := p
		if !filepath.IsAbs(path) {
			path = filepath.Join(config.ProjectDir, p)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			foundPath = true
			args = append(args, path)
		}
	}

	if !foundPath {
		c.Result.Status = result.Pass
		c.AddPass("no paths found to run phpstan on")
		return
	}

	c.DataMap = map[string][]byte{}
	c.DataMap["phpstan"], err = command.ShellCommander(phpstanPath, args...).Output()
	if err != nil {
		if pathErr, ok := err.(*fs.PathError); ok {
			c.AddBreach(&result.ValueBreach{
				ValueLabel: pathErr.Path,
				Value:      pathErr.Err.Error()})
		} else if len(c.DataMap["phpstan"]) == 0 { // If errors were found, exit code will be 1.
			c.AddBreach(&result.ValueBreach{
				ValueLabel: "Phpstan failed to run",
				Value:      string(err.(*exec.ExitError).Stderr)})
		}
	}
}

// HasData is overridden here to prevent the check from failing if there is no
// directory for phpstan to scan.
func (c *PhpStanCheck) HasData(failCheck bool) bool {
	if c.DataMap == nil && len(c.Result.Passes) == 0 {
		if failCheck {
			c.AddBreach(&result.ValueBreach{Value: "no data available"})
		}
		return false
	}
	return true
}

// UnmarshalDataMap parses the phpstan json into the PhpStan
// type for further processing.
func (c *PhpStanCheck) UnmarshalDataMap() {
	if c.Result.Status == result.Pass {
		return
	}

	if len(c.DataMap["phpstan"]) == 0 {
		c.Result.Status = result.Pass
		c.AddWarning("Unhandled PHPStan response, unable to determine status.")
		return
	}

	c.phpstanResult = PhpStanResult{}
	err := json.Unmarshal(c.DataMap["phpstan"], &c.phpstanResult)
	if err != nil {
		c.AddBreach(&result.ValueBreach{
			ValueLabel: "unable to parse phpstan result",
			Value:      err.Error()})
		return
	}

	// It's an empty slice, not a map, meaning no file errors.
	if string(c.phpstanResult.FilesRaw) == "[]" {
		return
	}

	// Unmarshal file errors.
	err = json.Unmarshal(c.phpstanResult.FilesRaw, &c.phpstanResult.Files)
	if err != nil {
		c.AddBreach(&result.ValueBreach{
			ValueLabel: "unable to parse phpstan file errors",
			Value:      err.Error()})
		return
	}
}

// RunCheck processes the parsed data and populates the errors, if any.
func (c *PhpStanCheck) RunCheck() {
	if c.phpstanResult.Totals.Errors == 0 && c.phpstanResult.Totals.FileErrors == 0 {
		c.AddPass("no error found")
		c.Result.Status = result.Pass
		return
	}

	for file, errors := range c.phpstanResult.Files {
		errLines := []string{}
		for _, er := range errors.Messages {
			errLines = append(errLines, fmt.Sprintf("line %d: %s", er.Line, er.Message))

		}
		c.AddBreach(&result.KeyValueBreach{
			Key:   fmt.Sprintf("file contains banned functions: %s", file),
			Value: strings.Join(errLines, "\n"),
		})
	}

	if len(c.phpstanResult.Errors) > 0 {
		c.AddBreach(&result.ValueBreach{
			ValueLabel: "errors encountered when running phpstan",
			Value:      strings.Join(c.phpstanResult.Errors, "\n")})
	}
}
