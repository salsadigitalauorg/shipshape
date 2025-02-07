package internal

import (
	"io"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type AnalyseTest struct {
	Name             string
	Input            fact.Facter
	Analyser         analyse.Analyser
	ExpectedBreaches []breach.Breach
}

// TestAnalyse is used to run test scenarios in test tables.
func TestAnalyse(t *testing.T, at AnalyseTest) {
	t.Helper()
	assert := assert.New(t)

	// Hide logging output.
	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	at.Input.Collect()
	at.Analyser.SetInput(at.Input)
	at.Analyser.Analyse()

	assert.Len(at.Analyser.GetResult().Breaches, len(at.ExpectedBreaches))
	assert.ElementsMatch(at.ExpectedBreaches, at.Analyser.GetResult().Breaches)
}
