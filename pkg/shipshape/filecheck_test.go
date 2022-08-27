package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestFileCheckMerge(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.FileCheck{
		CheckBase:         shipshape.CheckBase{Name: "filecheck1"},
		Path:              "file-initial",
		DisallowedPattern: "pattern-initial",
	}
	err := c.Merge(&shipshape.FileCheck{
		Path: "file-final",
	})
	assert.Nil(err)
	assert.EqualValues(shipshape.FileCheck{
		CheckBase:         shipshape.CheckBase{Name: "filecheck1"},
		Path:              "file-final",
		DisallowedPattern: "pattern-initial",
	}, c)

	err = c.Merge(&shipshape.FileCheck{
		DisallowedPattern: "pattern-final",
	})
	assert.Nil(err)
	assert.EqualValues(shipshape.FileCheck{
		CheckBase:         shipshape.CheckBase{Name: "filecheck1"},
		Path:              "file-final",
		DisallowedPattern: "pattern-final",
	}, c)

	err = c.Merge(&shipshape.FileCheck{
		CheckBase:         shipshape.CheckBase{Name: "filecheck2"},
		DisallowedPattern: "pattern-final",
	})
	assert.Error(err, "can only merge checks with the same name")
}

func TestFileCheckRunCheck(t *testing.T) {
	assert := assert.New(t)

	shipshape.ProjectDir = "testdata"
	c := shipshape.FileCheck{
		Path:              "file-non-existent",
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Init(shipshape.File)
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Equal(0, len(c.Result.Passes))
	assert.EqualValues(
		[]string{"lstat testdata/file-non-existent: no such file or directory"},
		c.Result.Failures,
	)

	c = shipshape.FileCheck{
		Path:              "file",
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Init(shipshape.File)
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Equal(0, len(c.Result.Passes))
	assert.EqualValues(
		[]string{
			"Illegal file found: testdata/file/adminer.php",
			"Illegal file found: testdata/file/sub/phpmyadmin.php",
		},
		c.Result.Failures,
	)

	c = shipshape.FileCheck{
		Path:              "file/correct",
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Init(shipshape.File)
	c.RunCheck()

	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.Equal(0, len(c.Result.Failures))
	assert.EqualValues([]string{"No illegal files"}, c.Result.Passes)
}
