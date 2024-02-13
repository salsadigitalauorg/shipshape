package internal

import (
	"io"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// CheckTest can be used to create test scenarios, especially using test tables,
// for the RunCheck method using TestRunCheck below.
type RunCheckTest struct {
	// Name of the test.
	Name  string
	Check config.Check
	// Initialise the check before testing.
	Init bool
	// Sort the results after running the check.
	Sort bool
	// Func to run before running the check
	PreRun func(t *testing.T)
	// Expected values after running the check.
	ExpectStatus result.Status
	ExpectNoPass bool
	ExpectPasses []string
	ExpectNoFail bool
	ExpectFails  []result.Breach
}

// TestRunCheck can be used to run test scenarios in test tables.
func TestRunCheck(t *testing.T, ctest RunCheckTest) {
	t.Helper()
	assert := assert.New(t)
	// Hide logging output.
	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	c := ctest.Check

	if ctest.Init {
		c.Init(c.GetType())
	}

	if ctest.PreRun != nil {
		ctest.PreRun(t)
	}
	c.RunCheck()

	r := c.GetResult()
	r.DetermineResultStatus(false)
	if ctest.Sort {
		r.Sort()
	}

	assert.Equal(ctest.ExpectStatus, r.Status)

	if ctest.ExpectNoPass {
		assert.Empty(r.Passes)
	} else {
		assert.ElementsMatchf(
			ctest.ExpectPasses,
			r.Passes,
			"Expected passes: %#v \nGot %#v", ctest.ExpectPasses, r.Passes)
	}

	if ctest.ExpectNoFail {
		assert.Empty(r.Breaches)
	} else {
		assert.ElementsMatchf(
			ctest.ExpectFails,
			r.Breaches,
			"Expected fails: %#v \nGot %#v", ctest.ExpectFails, r.Breaches)
	}
}
