package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

var cFalse = false
var cTrue = true

func TestYamlCheckMerge(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.YamlCheck{
		Path:           "path1",
		File:           "file1.yml",
		Pattern:        "pattern1",
		ExcludePattern: "excludePattern1",
		IgnoreMissing:  &cFalse,
	}
	c.Merge(&shipshape.YamlCheck{
		Path:  "path2",
		Files: []string{"slcFile1.yml", "slcFile2.yml"},
	})
	assert.EqualValues(shipshape.YamlCheck{
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

	mockCheck := func() shipshape.YamlCheck {
		return shipshape.YamlCheck{
			YamlBase: shipshape.YamlBase{
				Values: []shipshape.KeyValue{
					{Key: "check.interval_days", Value: "7"},
				},
			},
			Path: "yaml",
		}
	}

	c := mockCheck()
	c.FetchData()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.EqualValues([]string{"no file provided"}, c.Result.Failures)

	// Non-existent file.
	shipshape.ProjectDir = "testdata"
	c = mockCheck()
	c.Init(shipshape.Yaml)
	c.File = "non-existent.yml"
	c.FetchData()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.EqualValues([]string{"open testdata/yaml/non-existent.yml: no such file or directory"}, c.Result.Failures)

	// Non-existent file with ignore missing.
	c = mockCheck()
	c.File = "non-existent.yml"
	c.IgnoreMissing = &cTrue
	c.FetchData()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.EqualValues([]string{"File testdata/yaml/non-existent.yml does not exist"}, c.Result.Passes)

	// Single file.
	c = mockCheck()
	c.File = "update.settings.yml"
	c.FetchData()
	// Should not fail yet.
	assert.NotEqual(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.True(c.HasData(false))
	c.UnmarshalDataMap()
	c.RunCheck(false)
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.EqualValues([]string{"[yaml/update.settings.yml] 'check.interval_days' equals '7'"}, c.Result.Passes)

	// Bad File pattern.
	c = mockCheck()
	c.Pattern = "*.bar.yml"
	c.Path = ""
	c.FetchData()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.EqualValues([]string{"error parsing regexp: missing argument to repetition operator: `*`"}, c.Result.Failures)

	// File pattern with no matching files.
	c = mockCheck()
	c.Pattern = "bla.*.yml"
	c.FetchData()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.EqualValues([]string{"no matching config files found"}, c.Result.Failures)

	// File pattern with no matching files, ignoring missing.
	c = mockCheck()
	c.Pattern = "bla.*.yml"
	c.IgnoreMissing = &cTrue
	c.FetchData()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.Empty(c.Result.Failures)
	assert.EqualValues([]string{"no matching config files found"}, c.Result.Passes)

	// Correct single file pattern & value.
	c = mockCheck()
	c.Pattern = "foo.bar.yml"
	c.Path = "yaml/dir/subdir"
	c.FetchData()
	assert.NotEqual(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Failures)
	c.UnmarshalDataMap()
	c.RunCheck(false)
	assert.EqualValues([]string{"[testdata/yaml/dir/subdir/foo.bar.yml] 'check.interval_days' equals '7'"}, c.Result.Passes)
	assert.Empty(c.Result.Failures)

	// Recursive file lookup.
	c = mockCheck()
	c.Pattern = ".*.bar.yml"
	c.FetchData()
	assert.NotEqual(shipshape.Fail, c.Result.Status)
	assert.Empty(c.Result.Failures)
	c.UnmarshalDataMap()
	c.RunCheck(false)
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.ElementsMatch(
		[]string{
			"[testdata/yaml/dir/foo.bar.yml] 'check.interval_days' equals '7'",
			"[testdata/yaml/dir/subdir/foo.bar.yml] 'check.interval_days' equals '7'",
			"[testdata/yaml/foo.bar.yml] 'check.interval_days' equals '7'"},
		c.Result.Passes)
	assert.ElementsMatch(
		[]string{
			"[testdata/yaml/dir/subdir/zoom.bar.yml] 'check.interval_days' equals '5', expected '7'",
			"[testdata/yaml/dir/zoom.bar.yml] 'check.interval_days' equals '5', expected '7'",
			"[testdata/yaml/zoom.bar.yml] 'check.interval_days' equals '5', expected '7'"},
		c.Result.Failures)
}
