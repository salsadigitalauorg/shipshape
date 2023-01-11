package shipshape_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestCheckBaseInit(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.CheckBase{Name: "foo"}
	assert.Equal("foo", c.GetName())

	c.Init(shipshape.File)
	assert.Equal(shipshape.NormalSeverity, c.Severity)
	assert.Equal("foo", c.Result.Name)
	assert.Equal(shipshape.NormalSeverity, c.Result.Severity)
	assert.Equal(shipshape.File, c.GetType())
}

func TestCheckBaseMerge(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.CheckBase{Name: "foo"}
	err := c.Merge(&shipshape.CheckBase{Name: "bar"})
	assert.Equal(fmt.Errorf("can only merge checks with the same name"), err)

	c = shipshape.CheckBase{Name: "foo", Severity: shipshape.HighSeverity}
	c.Merge(&shipshape.CheckBase{Name: "foo"})
	assert.Equal(shipshape.HighSeverity, c.Severity)

	c = shipshape.CheckBase{Severity: shipshape.LowSeverity}
	c.Merge(&shipshape.CheckBase{Name: "foo"})
	assert.Equal(shipshape.LowSeverity, c.Severity)
}

func TestCheckBaseRunCheck(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.CheckBase{}
	c.FetchData()
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues([]string{"not implemented"}, c.Result.Failures)
}

type testCheckRemediationNotSupported struct {
	shipshape.YamlBase `yaml:",inline"`
}

type testCheckRemediationSupported struct {
	shipshape.YamlBase `yaml:",inline"`
}

func (c *testCheckRemediationSupported) SupportsRemediation() bool {
	return true
}

func (c *testCheckRemediationSupported) Remediate() error {
	return errors.New("foo")
}

func TestSupportsRemediation(t *testing.T) {
	assert := assert.New(t)

	t.Run("notSupported", func(t *testing.T) {
		c := testCheckRemediationNotSupported{}
		assert.False(c.SupportsRemediation())

		err := c.Remediate()
		assert.NoError(err)
		assert.Empty(c.Result.Passes)
		assert.Empty(c.Result.Failures)
		assert.ElementsMatch(
			c.Result.Warnings,
			[]string{"This check does not currently implement remediation."})
	})

	t.Run("supported", func(t *testing.T) {
		c := testCheckRemediationSupported{}
		assert.True(c.SupportsRemediation())

		err := c.Remediate()
		assert.EqualError(err, "foo")
		assert.Empty(c.Result.Passes)
		assert.Empty(c.Result.Warnings)
		assert.Empty(c.Result.Failures)
	})
}
