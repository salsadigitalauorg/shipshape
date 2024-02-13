package file_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/file"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestFileCheckMerge(t *testing.T) {
	assert := assert.New(t)

	c := FileCheck{
		CheckBase:         config.CheckBase{Name: "filecheck1"},
		Path:              "file-initial",
		DisallowedPattern: "pattern-initial",
	}
	err := c.Merge(&FileCheck{
		Path: "file-final",
	})
	assert.Nil(err)
	assert.EqualValues(FileCheck{
		CheckBase:         config.CheckBase{Name: "filecheck1"},
		Path:              "file-final",
		DisallowedPattern: "pattern-initial",
	}, c)

	err = c.Merge(&FileCheck{
		DisallowedPattern: "pattern-final",
	})
	assert.Nil(err)
	assert.EqualValues(FileCheck{
		CheckBase:         config.CheckBase{Name: "filecheck1"},
		Path:              "file-final",
		DisallowedPattern: "pattern-final",
	}, c)

	err = c.Merge(&FileCheck{
		CheckBase:         config.CheckBase{Name: "filecheck2"},
		DisallowedPattern: "pattern-final",
	})
	assert.Error(err, "can only merge checks with the same name")
}

func TestFileCheckRunCheck(t *testing.T) {
	assert := assert.New(t)

	config.ProjectDir = "testdata"
	c := FileCheck{
		Path:              "file-non-existent",
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Name = "filecheck1"
	c.Init(File)
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.Equal(0, len(c.Result.Passes))
	assert.EqualValues(
		[]result.Breach{&result.ValueBreach{
			CheckType:  "file",
			CheckName:  "filecheck1",
			BreachType: result.BreachTypeValue,
			Severity:   "normal",
			ValueLabel: "error finding files",
			Value:      "lstat testdata/file-non-existent: no such file or directory",
		}},
		c.Result.Breaches,
	)

	c = FileCheck{
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Name = "filecheck2"
	c.Init(File)
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.Equal(0, len(c.Result.Passes))
	assert.EqualValues(
		[]result.Breach{
			&result.ValueBreach{
				BreachType: "value",
				CheckType:  "file",
				CheckName:  "filecheck2",
				Severity:   "normal",
				ValueLabel: "filecheck2 - illegal files found",
				Value:      "testdata/adminer.php\ntestdata/sub/phpmyadmin.php",
			},
		},
		c.Result.Breaches,
	)

	c = FileCheck{
		Path:              "correct",
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Init(File)
	c.RunCheck()

	assert.Equal(result.Pass, c.Result.Status)
	assert.Equal(0, len(c.Result.Breaches))
	assert.EqualValues([]string{"No illegal files"}, c.Result.Passes)
}
