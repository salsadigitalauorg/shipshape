package file_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/file"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
)

var cTrue = true

func TestFileDiffCheck_Merge(t *testing.T) {
	assertions := assert.New(t)

	c := file.FileDiffCheck{
		CheckBase:  config.CheckBase{Name: "filediffcheck1"},
		SourceFile: "source-initial",
		TargetFile: "target-initial",
	}
	err := c.Merge(&file.FileDiffCheck{
		SourceFile: "source-final",
		TargetFile: "target-final",
	})
	assertions.Nil(err)
	assertions.EqualValues(file.FileDiffCheck{
		CheckBase:  config.CheckBase{Name: "filediffcheck1"},
		SourceFile: "source-final",
		TargetFile: "target-final",
	}, c)

	c = file.FileDiffCheck{
		CheckBase:    config.CheckBase{Name: "filediffcheck2"},
		SourceFile:   "source-initial",
		TargetFile:   "target-initial",
		ContextLines: 0,
	}
	err = c.Merge(&file.FileDiffCheck{
		ContextLines: 1,
	})
	assertions.Nil(err)
	assertions.EqualValues(file.FileDiffCheck{
		CheckBase:    config.CheckBase{Name: "filediffcheck2"},
		SourceFile:   "source-initial",
		TargetFile:   "target-initial",
		ContextLines: 1,
	}, c)

	c = file.FileDiffCheck{
		CheckBase:     config.CheckBase{Name: "filediffcheck3"},
		SourceFile:    "source-initial",
		TargetFile:    "target-initial",
		SourceContext: map[string]any{"key1": "value1"},
	}
	err = c.Merge(&file.FileDiffCheck{
		SourceContext: map[string]any{"key1": "value2"},
	})
	assertions.Nil(err)
	assertions.EqualValues(file.FileDiffCheck{
		CheckBase:     config.CheckBase{Name: "filediffcheck3"},
		SourceFile:    "source-initial",
		TargetFile:    "target-initial",
		SourceContext: map[string]any{"key1": "value2"},
	}, c)
}

func TestFileDiffCheck_FetchData(t *testing.T) {
	assertions := assert.New(t)

	config.ProjectDir = "testdata/filediff/"

	t.Run("failOnNoSource", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:  config.CheckBase{Name: "filediffcheck"},
			TargetFile: "file1.txt",
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Equal(0, len(c.Result.Passes))
		assertions.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				CheckType:  "filediff",
				CheckName:  "filediffcheck",
				BreachType: breach.BreachTypeValue,
				Severity:   "normal",
				Value:      "no source file provided",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("failOnNoTarget", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:  config.CheckBase{Name: "filediffcheck"},
			SourceFile: "file1.txt",
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Equal(0, len(c.Result.Passes))
		assertions.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				CheckType:  "filediff",
				CheckName:  "filediffcheck",
				BreachType: breach.BreachTypeValue,
				Severity:   "normal",
				Value:      "no target file provided",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("failOnSourceNotExist", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:  config.CheckBase{Name: "filediffcheck1"},
			SourceFile: "file0.txt",
			TargetFile: "file1.txt",
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Equal(0, len(c.Result.Passes))
		assertions.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				CheckType:  "filediff",
				CheckName:  "filediffcheck1",
				BreachType: breach.BreachTypeValue,
				Severity:   "normal",
				ValueLabel: "error fetching source file: file0.txt",
				Value:      "open testdata/filediff/file0.txt: no such file or directory",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("failOnTargetNotExist", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:  config.CheckBase{Name: "filediffcheck1"},
			SourceFile: "file1.txt",
			TargetFile: "file0.txt",
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Equal(0, len(c.Result.Passes))
		assertions.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				CheckType:  "filediff",
				CheckName:  "filediffcheck1",
				BreachType: breach.BreachTypeValue,
				Severity:   "normal",
				ValueLabel: "error reading target file: file0.txt",
				Value:      "open testdata/filediff/file0.txt: no such file or directory",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("passOnIgnoreMissingTarget", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:     config.CheckBase{Name: "filediffcheck1"},
			SourceFile:    "file1.txt",
			TargetFile:    "file0.txt",
			IgnoreMissing: &cTrue,
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Pass, c.Result.Status)
		assertions.Equal(0, len(c.Result.Breaches))
		assertions.EqualValues([]string{"Target file file0.txt does not exist"}, c.Result.Passes)
	})

	t.Run("failOnMalformedJinjaSource", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:     config.CheckBase{Name: "filediffcheck1"},
			SourceFile:    "file3.txt",
			TargetFile:    "file2.txt",
			SourceContext: map[string]any{"VERSION": 1},
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Equal(0, len(c.Result.Passes))
		assertions.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				CheckType:  "filediff",
				CheckName:  "filediffcheck1",
				BreachType: breach.BreachTypeValue,
				Severity:   "normal",
				ValueLabel: "error parsing source file: file3.txt",
				Value:      "failed to parse template 'This is file #{{ VERSION }.\n': '}}' expected here (Line: 0 Col: 0, near \"Unexpected delimiter \"}\"\")",
			}},
			c.Result.Breaches,
		)
	})
}

func TestFileDiffCheck_RunCheck(t *testing.T) {
	assertions := assert.New(t)

	config.ProjectDir = "testdata/filediff/"

	t.Run("passOnIdenticalFiles", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:  config.CheckBase{Name: "filediffcheck"},
			SourceFile: "file1.txt",
			TargetFile: "file1.txt",
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Pass, c.Result.Status)
		assertions.Equal(0, len(c.Result.Breaches))
		assertions.EqualValues([]string{"Target file file1.txt is identical to Source file file1.txt"}, c.Result.Passes)
	})

	t.Run("failOnDifferentFiles", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:  config.CheckBase{Name: "filediffcheck"},
			SourceFile: "file1.txt",
			TargetFile: "file2.txt",
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Equal(0, len(c.Result.Passes))
		assertions.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				CheckType:  "filediff",
				CheckName:  "filediffcheck",
				BreachType: breach.BreachTypeValue,
				Severity:   "normal",
				ValueLabel: "Target file file2.txt is different from Source file file1.txt",
				Value:      "diff: \n--- file1.txt\n+++ file2.txt\n@@ -1 +1 @@\n-This is file #1.\n+This is file #2.\n",
			}},
			c.Result.Breaches,
		)
	})

	t.Run("passOnIdenticalJinjaFiles", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:     config.CheckBase{Name: "filediffcheck"},
			SourceFile:    "file4.txt",
			TargetFile:    "file1.txt",
			SourceContext: map[string]any{"VERSION": 1},
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Pass, c.Result.Status)
		assertions.Equal(0, len(c.Result.Breaches))
		assertions.EqualValues([]string{"Target file file1.txt is identical to Source file file4.txt"}, c.Result.Passes)
	})

	t.Run("failOnDifferentJinjaFiles", func(t *testing.T) {
		c := file.FileDiffCheck{
			CheckBase:     config.CheckBase{Name: "filediffcheck"},
			SourceFile:    "file4.txt",
			TargetFile:    "file2.txt",
			SourceContext: map[string]any{"VERSION": 1},
		}
		c.Init(file.FileDiff)
		c.FetchData()
		c.RunCheck()
		c.Result.DetermineResultStatus(false)
		assertions.Equal(result.Fail, c.Result.Status)
		assertions.Equal(0, len(c.Result.Passes))
		assertions.EqualValues(
			[]breach.Breach{&breach.ValueBreach{
				CheckType:  "filediff",
				CheckName:  "filediffcheck",
				BreachType: breach.BreachTypeValue,
				Severity:   "normal",
				ValueLabel: "Target file file2.txt is different from Source file file4.txt",
				Value:      "diff: \n--- file4.txt\n+++ file2.txt\n@@ -1 +1 @@\n-This is file #1.\n+This is file #2.\n",
			}},
			c.Result.Breaches,
		)
	})
}
