package internal

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

// CheckTest can be used to create test scenarios, especially using test tables,
// for the RunCheck method using TestRunCheck below.
type RunCheckTest struct {
	// Name of the test.
	Name  string
	Check shipshape.Check
	// Initialise the check before testing.
	Init bool
	// Run the check with Remediate flag.
	Remediate bool
	// Sort the results after running the check.
	Sort bool
	// Func to run before running the check
	PreRun func(t *testing.T)
	// Expected values after running the check.
	ExpectStatus         shipshape.CheckStatus
	ExpectNoPass         bool
	ExpectPasses         []string
	ExpectNoFail         bool
	ExpectFails          []string
	ExpectNoRemediations bool
	ExpectRemediations   []string
}

// TestRunCheck can be used to run test scenarios in test tables.
func TestRunCheck(t *testing.T, ctest RunCheckTest) {
	t.Helper()
	assert := assert.New(t)

	c := ctest.Check

	if ctest.Init {
		c.Init(c.GetType())
	}

	if ctest.PreRun != nil {
		ctest.PreRun(t)
	}
	c.RunCheck(ctest.Remediate)

	r := c.GetResult()
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
		assert.Empty(r.Failures)
	} else {
		assert.ElementsMatchf(
			ctest.ExpectFails,
			r.Failures,
			"Expected fails: %#v \nGot %#v", ctest.ExpectFails, r.Failures)
	}

	if ctest.ExpectNoRemediations {
		assert.Empty(r.Remediations)
	} else {
		assert.ElementsMatchf(
			ctest.ExpectRemediations,
			r.Remediations,
			"Expected remediations: %#v \nGot %#v", ctest.ExpectRemediations, r.Remediations)
	}
}
