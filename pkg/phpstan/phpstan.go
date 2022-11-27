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

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

//go:generate go run ../../cmd/gen.go registry --checkpackage=phpstan

const PhpStan shipshape.CheckType = "phpstan"

type PhpStanCheck struct {
	shipshape.CheckBase `yaml:",inline"`
	Bin                 string   `yaml:"binary"`
	Config              string   `yaml:"configuration"`
	Paths               []string `yaml:"paths"`
	phpstanResult       PhpStanResult
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
	shipshape.ChecksRegistry[PhpStan] = func() shipshape.Check { return &PhpStanCheck{} }
}

func init() {
	RegisterChecks()
}

const PhpstanDefaultPath = "vendor/phpstan/phpstan/phpstan"

var ExecCommand = exec.Command

// Merge implementation for file check.
func (c *PhpStanCheck) Merge(mergeCheck shipshape.Check) error {
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
		path = filepath.Join(shipshape.ProjectDir, PhpstanDefaultPath)
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
		configPath = filepath.Join(shipshape.ProjectDir, configPath)
	}

	args := []string{
		"analyse",
		fmt.Sprintf("--configuration=%s", configPath),
		"--no-progress",
		"--error-format=json",
	}
	for _, p := range c.Paths {
		path := p
		if !filepath.IsAbs(path) {
			path = filepath.Join(shipshape.ProjectDir, p)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			args = append(args, path)
		}
	}

	c.DataMap = map[string][]byte{}
	c.DataMap["phpstan"], err = ExecCommand(phpstanPath, args...).Output()
	if err != nil {
		if pathErr, ok := err.(*fs.PathError); ok {
			c.AddFail(pathErr.Path + ": " + pathErr.Err.Error())
		} else if len(c.DataMap["phpstan"]) == 0 { // If errors were found, exit code will be 1.
			c.AddFail("Phpstan failed to run: " + string(err.(*exec.ExitError).Stderr))
		}
	}
}

// UnmarshalDataMap parses the phpstan json into the PhpStan
// type for further processing.
func (c *PhpStanCheck) UnmarshalDataMap() {
	if len(c.DataMap["phpstan"]) == 0 {
		c.AddFail("no data provided")
		return
	}

	c.phpstanResult = PhpStanResult{}
	err := json.Unmarshal(c.DataMap["phpstan"], &c.phpstanResult)
	if err != nil {
		c.AddFail(err.Error())
		return
	}

	// It's an empty slice, not a map, meaning no file errors.
	if string(c.phpstanResult.FilesRaw) == "[]" {
		return
	}

	// Unmarshal file errors.
	err = json.Unmarshal(c.phpstanResult.FilesRaw, &c.phpstanResult.Files)
	if err != nil {
		c.AddFail(err.Error())
		return
	}
}

// RunCheck processes the parsed data and populates the errors, if any.
func (c *PhpStanCheck) RunCheck() {
	if c.phpstanResult.Totals.Errors == 0 && c.phpstanResult.Totals.FileErrors == 0 {
		c.AddPass("no error found")
		c.Result.Status = shipshape.Pass
		return
	}

	for file, errors := range c.phpstanResult.Files {
		for _, er := range errors.Messages {
			c.AddFail(fmt.Sprintf("[%s] Line %d: %s", file, er.Line, er.Message))
		}
	}

	for _, er := range c.phpstanResult.Errors {
		c.AddFail(er)
	}
}
