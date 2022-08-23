package shipshape_test

import (
	"fmt"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestFileCheckMerge(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.FileCheck{
		Path:              "file-initial",
		DisallowedPattern: "pattern-initial",
	}
	c.Init("", shipshape.File)

	c2 := shipshape.YamlCheck{}
	c2.Init("", shipshape.Yaml)
	err := c.Merge(&c2)
	assert.Equal(fmt.Errorf("can only merge checks of the same type"), err)

	c = shipshape.FileCheck{
		Path:              "file-initial",
		DisallowedPattern: "pattern-initial",
	}
	c.Merge(&shipshape.FileCheck{
		Path: "file-final",
	})
	assert.EqualValues(shipshape.FileCheck{
		Path:              "file-final",
		DisallowedPattern: "pattern-initial",
	}, c)

	c.Merge(&shipshape.FileCheck{
		DisallowedPattern: "pattern-final",
	})
	assert.EqualValues(shipshape.FileCheck{
		Path:              "file-final",
		DisallowedPattern: "pattern-final",
	}, c)
}

func TestFileCheckRunCheck(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.FileCheck{
		Path:              "file-non-existent",
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Init("testdata", shipshape.File)
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
	c.Init("testdata", shipshape.File)
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
	c.Init("testdata", shipshape.File)
	c.RunCheck()

	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.Equal(0, len(c.Result.Failures))
	assert.EqualValues([]string{"No illegal files"}, c.Result.Passes)
}
