package yaml_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/stretchr/testify/assert"
)

var cFalse = false
var cTrue = true

func TestYamlCheckMerge(t *testing.T) {
	assert := assert.New(t)

	c := YamlCheck{
		Path:           "path1",
		File:           "file1.yml",
		Pattern:        "pattern1",
		ExcludePattern: "excludePattern1",
		IgnoreMissing:  &cFalse,
	}
	c.Merge(&YamlCheck{
		Path:  "path2",
		Files: []string{"slcFile1.yml", "slcFile2.yml"},
	})
	assert.EqualValues(YamlCheck{
		Path:           "path2",
		File:           "file1.yml",
		Files:          []string{"slcFile1.yml", "slcFile2.yml"},
		Pattern:        "pattern1",
		ExcludePattern: "excludePattern1",
		IgnoreMissing:  &cFalse,
	}, c)
}

func TestYamlCheck(t *testing.T) {
	assert := assert.New(t)

	mockCheck := func() YamlCheck {
		return YamlCheck{
			YamlBase: YamlBase{
				Values: []KeyValue{
					{Key: "check.interval_days", Value: "7"},
				},
			},
		}
	}

	c := mockCheck()
	c.FetchData()
	assert.Equal(config.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.EqualValues([]string{"no file provided"}, c.Result.Failures)

	// Non-existent file.
	config.ProjectDir = "testdata"
	c = mockCheck()
	c.Init(Yaml)
	c.File = "non-existent.yml"
	c.FetchData()
	assert.Equal(config.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.EqualValues([]string{"open testdata/non-existent.yml: no such file or directory"}, c.Result.Failures)

	// Non-existent file with ignore missing.
	c = mockCheck()
	c.File = "non-existent.yml"
	c.IgnoreMissing = &cTrue
	c.FetchData()
	assert.Equal(config.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.EqualValues([]string{"File testdata/non-existent.yml does not exist"}, c.Result.Passes)

	// Single file.
	c = mockCheck()
	c.File = "update.settings.yml"
	c.FetchData()
	// Should not fail yet.
	assert.NotEqual(config.Fail, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.True(c.HasData(false))
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(config.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.EqualValues([]string{"[update.settings.yml] 'check.interval_days' equals '7'"}, c.Result.Passes)

	// Bad File pattern.
	c = mockCheck()
	c.Pattern = "*.bar.yml"
	c.Path = ""
	c.FetchData()
	assert.Equal(config.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.EqualValues([]string{"error parsing regexp: missing argument to repetition operator: `*`"}, c.Result.Failures)

	// File pattern with no matching files.
	c = mockCheck()
	c.Pattern = "bla.*.yml"
	c.FetchData()
	assert.Equal(config.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.EqualValues([]string{"no matching config files found"}, c.Result.Failures)

	// File pattern with no matching files, ignoring missing.
	c = mockCheck()
	c.Pattern = "bla.*.yml"
	c.IgnoreMissing = &cTrue
	c.FetchData()
	assert.Equal(config.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.EqualValues([]string{"no matching config files found"}, c.Result.Passes)

	// Correct single file pattern & value.
	c = mockCheck()
	c.Pattern = "foo.bar.yml"
	c.Path = "dir/subdir"
	c.FetchData()
	assert.NotEqual(config.Fail, c.Result.Status)
	assert.Empty(c.Result.Failures)
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.EqualValues([]string{"[testdata/dir/subdir/foo.bar.yml] 'check.interval_days' equals '7'"}, c.Result.Passes)
	assert.Empty(c.Result.Failures)

	// Recursive file lookup.
	c = mockCheck()
	c.Pattern = ".*.bar.yml"
	c.FetchData()
	assert.NotEqual(config.Fail, c.Result.Status)
	assert.Empty(c.Result.Failures)
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(config.Fail, c.Result.Status)
	assert.ElementsMatch(
		[]string{
			"[testdata/dir/foo.bar.yml] 'check.interval_days' equals '7'",
			"[testdata/dir/subdir/foo.bar.yml] 'check.interval_days' equals '7'",
			"[testdata/foo.bar.yml] 'check.interval_days' equals '7'"},
		c.Result.Passes)
	assert.ElementsMatch(
		[]string{
			"[testdata/dir/subdir/zoom.bar.yml] 'check.interval_days' equals '5', expected '7'",
			"[testdata/dir/zoom.bar.yml] 'check.interval_days' equals '5', expected '7'",
			"[testdata/zoom.bar.yml] 'check.interval_days' equals '5', expected '7'"},
		c.Result.Failures)
}
