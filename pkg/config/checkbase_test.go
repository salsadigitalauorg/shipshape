package config_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/config/testdata/testchecks"
	"github.com/salsadigitalauorg/shipshape/pkg/result"

	"github.com/stretchr/testify/assert"
)

const testCheckForCheckBaseInitType CheckType = "testCheckForCheckBaseInitType"

func TestCheckBaseInit(t *testing.T) {
	assert := assert.New(t)

	c := CheckBase{Name: "foo"}
	assert.Equal("foo", c.GetName())

	c.Init(testCheckForCheckBaseInitType)
	assert.Equal(NormalSeverity, c.Severity)
	assert.Equal("foo", c.Result.Name)
	assert.Equal(string(NormalSeverity), c.Result.Severity)
	assert.Equal(testCheckForCheckBaseInitType, c.GetType())
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

func TestRequiresData(t *testing.T) {
	assert := assert.New(t)

	var c Check
	c = &CheckBase{Name: "foo"}
	assert.True(c.RequiresData())

	c = &testchecks.TestCheck1Check{}
	c.Init(testchecks.TestCheck1)
	assert.False(c.RequiresData())
}

func TestHasData(t *testing.T) {
	assert := assert.New(t)

	c := CheckBase{Name: "foo"}
	assert.False(c.HasData(false))
	assert.Empty(c.Result.Breaches)
	c.Result.DetermineResultStatus(false)
	assert.NotEqual(result.Fail, c.Result.Status)

	assert.False(c.HasData(true))
	assert.EqualValues([]breach.Breach{
		&breach.ValueBreach{
			BreachType: "value",
			CheckName:  "foo",
			Value:      "no data available",
		},
	}, c.Result.Breaches)
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)

	c = CheckBase{Name: "foo", DataMap: map[string][]byte{"foo": []byte(`bar`)}}
	assert.True(c.HasData(true))
	assert.Empty(c.Result.Breaches)
	c.Result.DetermineResultStatus(false)
	assert.NotEqual(result.Fail, c.Result.Status)
}

func TestAddBreach(t *testing.T) {
	assert := assert.New(t)

	const vbCheckType CheckType = "vbCheckType"
	const kvbCheckType CheckType = "kvbCheckType"
	const kvsbCheckType CheckType = "kvsbCheckType"

	tests := []struct {
		name      string
		checkName string
		checkType CheckType
		severity  Severity
		breach    breach.Breach
	}{
		{
			name:      "ValueBreach",
			checkType: vbCheckType,
			checkName: "vbCheck",
			severity:  "high",
			breach:    &breach.ValueBreach{},
		},
		{
			name:      "KeyValueBreach",
			checkType: kvbCheckType,
			checkName: "kvbCheck",
			severity:  "low",
			breach:    &breach.KeyValueBreach{},
		},
		{
			name:      "KeyValuesBreach",
			checkType: kvsbCheckType,
			checkName: "kvsbCheck",
			severity:  "normal",
			breach:    &breach.KeyValuesBreach{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := CheckBase{Name: test.checkName, Severity: test.severity}
			c.Init(test.checkType)
			c.AddBreach(test.breach)
			assert.Equal(string(test.checkType), c.Result.Breaches[0].GetCheckType())
			assert.Equal(test.checkName, c.Result.Breaches[0].GetCheckName())
			assert.Equal(string(test.severity), c.Result.Breaches[0].GetSeverity())
		})
	}
}

func TestAddPass(t *testing.T) {
	assert := assert.New(t)

	c := CheckBase{Name: "foo"}
	c.AddPass("with flying colours!")
	assert.EqualValues(result.Result{Passes: []string{"with flying colours!"}}, c.Result)
}

func TestAddWarning(t *testing.T) {
	assert := assert.New(t)

	c := CheckBase{Name: "foo"}
	c.AddWarning("not feeling great")
	assert.EqualValues(result.Result{Warnings: []string{"not feeling great"}}, c.Result)
}

func TestCheckBaseRunCheck(t *testing.T) {
	assert := assert.New(t)

	c := CheckBase{}
	c.FetchData()
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues(
		[]breach.Breach{&breach.ValueBreach{
			BreachType: breach.BreachTypeValue,
			Value:      "not implemented",
		}},
		c.Result.Breaches)

}

type testCheckRemediationNotSupported struct{ CheckBase }

type testCheckRemediationSupported struct{ CheckBase }

func (c *testCheckRemediationSupported) Remediate(interface{}) error {
	return errors.New("foo")
}

func TestRemediate(t *testing.T) {
	assert := assert.New(t)

	t.Run("notSupported", func(t *testing.T) {
		c := testCheckRemediationNotSupported{}

		c.Remediate()
		assert.Empty(c.Result.Passes)
		assert.Empty(c.Result.Breaches)
	})

	t.Run("supported", func(t *testing.T) {
		c := testCheckRemediationSupported{}

		err := c.Remediate(nil)
		assert.EqualError(err, "foo")
		assert.Empty(c.Result.Passes)
		assert.Empty(c.Result.Breaches)
	})
}
