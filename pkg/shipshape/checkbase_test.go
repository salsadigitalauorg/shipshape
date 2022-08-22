package shipshape_test

import (
	"fmt"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestCheckBaseInit(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.CheckBase{Name: "foo"}
	assert.Equal("foo", c.GetName())

	c.Init("baz", shipshape.File)
	assert.Equal("baz", shipshape.ProjectDir)
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

	c.Init("some-dir", shipshape.File)
	c2 := shipshape.CheckBase{Name: "bar"}
	c2.Init("some-dir", shipshape.Yaml)
	err = c.Merge(&c2)
	assert.Equal(fmt.Errorf("can only merge checks of the same type"), err)

	c = shipshape.CheckBase{Name: "foo", Severity: shipshape.HighSeverity}
	c.Merge(&shipshape.CheckBase{Name: "foo"})
	assert.Equal(shipshape.HighSeverity, c.Severity)
}

func TestCheckBaseRunCheck(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.CheckBase{}
	c.FetchData()
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues([]string{"not implemented"}, c.Result.Failures)
}
