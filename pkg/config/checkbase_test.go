package config_test

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/file"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"

	"github.com/stretchr/testify/assert"
)

func TestCheckBaseInit(t *testing.T) {
	assert := assert.New(t)

	c := CheckBase{Name: "foo"}
	assert.Equal("foo", c.GetName())

	c.Init(file.File)
	assert.Equal(NormalSeverity, c.Severity)
	assert.Equal("foo", c.Result.Name)
	assert.Equal(NormalSeverity, c.Result.Severity)
	assert.Equal(file.File, c.GetType())
}

func TestCheckBaseMerge(t *testing.T) {
	assert := assert.New(t)

	c := CheckBase{Name: "foo"}
	err := c.Merge(&CheckBase{Name: "bar"})
	assert.Equal(fmt.Errorf("can only merge checks with the same name"), err)

	c = CheckBase{Name: "foo", Severity: HighSeverity}
	c.Merge(&CheckBase{Name: "foo"})
	assert.Equal(HighSeverity, c.Severity)

	c = CheckBase{Severity: LowSeverity}
	c.Merge(&CheckBase{Name: "foo"})
	assert.Equal(LowSeverity, c.Severity)
}

func TestCheckBaseRunCheck(t *testing.T) {
	assert := assert.New(t)

	c := CheckBase{}
	c.FetchData()
	c.RunCheck()
	assert.Equal(Fail, c.Result.Status)
	assert.EqualValues([]string{"not implemented"}, c.Result.Failures)
}

type testCheckRemediationNotSupported struct {
	shipshape.YamlBase `yaml:",inline"`
}

type testCheckRemediationSupported struct {
	shipshape.YamlBase `yaml:",inline"`
}

func (c *testCheckRemediationSupported) Remediate(interface{}) error {
	return errors.New("foo")
}

func TestRemediate(t *testing.T) {
	assert := assert.New(t)

	t.Run("notSupported", func(t *testing.T) {
		c := testCheckRemediationNotSupported{}

		err := c.Remediate(nil)
		assert.NoError(err)
		assert.Empty(c.Result.Passes)
		assert.Empty(c.Result.Failures)
	})

	t.Run("supported", func(t *testing.T) {
		c := testCheckRemediationSupported{}

		err := c.Remediate(nil)
		assert.EqualError(err, "foo")
		assert.Empty(c.Result.Passes)
		assert.Empty(c.Result.Failures)
	})
}
