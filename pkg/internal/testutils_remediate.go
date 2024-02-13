package internal

import (
	"io"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// CheckTest can be used to create test scenarios, especially using test tables,
// for the RunCheck method using TestRunCheck below.
type RemediateTest struct {
	// Name of the test.
	Name     string
	Check    config.Check
	Breaches []result.Breach
	// Func to run before running Remediate
	PreRun func(t *testing.T)
	// Expected values after running Remediate.
	ExpectGeneratedCommand  string
	ExpectStatusFail        bool
	ExpectNoBreach          bool
	ExpectBreaches          []result.Breach
	ExpectRemediationStatus result.RemediationStatus
	ExpectNoRemediations    bool
	ExpectRemediations      []string
}

// TestRunCheck can be used to run test scenarios in test tables.
func TestRemediate(t *testing.T, rt RemediateTest) {
	t.Helper()
	assert := assert.New(t)
	// Hide logging output.
	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	if rt.PreRun != nil {
		rt.PreRun(t)
	}

	var generatedCommand string
	if rt.ExpectGeneratedCommand != "" {
		curShellCommander := command.ShellCommander
		defer func() { command.ShellCommander = curShellCommander }()
		command.ShellCommander = ShellCommanderMaker(nil, nil, &generatedCommand)
	}
	rt.Check.Remediate()
	if rt.ExpectGeneratedCommand != "" {
		assert.Equal(rt.ExpectGeneratedCommand, generatedCommand)
	}

	r := rt.Check.GetResult()
	r.DetermineResultStatus(true)

	if rt.ExpectStatusFail {
		assert.Equal(result.Fail, r.Status)
	}
	if rt.ExpectNoBreach {
		assert.Empty(r.Breaches)
	} else {
		assert.ElementsMatchf(
			rt.ExpectBreaches,
			r.Breaches,
			"Expected breaches: %#v \nGot %#v", rt.ExpectBreaches, r.Breaches)
	}

	assert.Equal(rt.ExpectRemediationStatus, r.RemediationStatus)
	if rt.ExpectNoRemediations {
		assert.NotEmpty(r.Breaches)
		remediationsFound := false
		for _, b := range r.Breaches {
			if b.GetRemediation().Status != "" {
				remediationsFound = true
				break
			}
		}
		assert.False(remediationsFound, "Expected no remediations, but found some")
	} else if len(rt.ExpectRemediations) > 0 {
		assert.NotEmpty(r.Breaches)
		remediationMsgs := []string{}
		for _, b := range r.Breaches {
			remediationMsgs = append(remediationMsgs, b.GetRemediation().Messages...)
		}
		assert.ElementsMatchf(
			rt.ExpectRemediations,
			remediationMsgs,
			"Expected remediations: %#v \nGot %#v", rt.ExpectRemediations, remediationMsgs)
	}
}
