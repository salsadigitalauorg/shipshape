package phpstan_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/phpstan"
	"github.com/stretchr/testify/assert"
)

func fakeExecCommand(command string, args ...string) bool {
	return true
}

func TestBinPathProvided(t *testing.T) {
	assert := assert.New(t)
	c := phpstan.PhpStanCheck{
		Bin:    "/my/custom/path/phpstan",
		Config: "/path/to/config",
	}

	assert.Equal(c.GetBinary(), "/my/custom/path/phpstan")
}

func TestBinPathDefault(t *testing.T) {
	assert := assert.New(t)
	c := phpstan.PhpStanCheck{
		Config: "/path/to/config",
	}

	assert.Equal(c.GetBinary(), "vendor/phpstan/phpstan/phpstan")
}
