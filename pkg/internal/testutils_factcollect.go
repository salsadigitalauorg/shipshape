package internal

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/fact/testdata"
)

type FactInputTest struct {
	DataFormat data.DataFormat
	Data       any
}

type FactCollectTest struct {
	Name   string
	Facter fact.Facter

	TestInput          FactInputTest
	ExpectedInputError error

	TestAdditionalInputs         map[string]FactInputTest
	ExpectedAdditionalInputsErrs []error

	ExpectedErrors []error
	ExpectedData   interface{}
}

// TestFactCollect is used to run test scenarios in test tables.
func TestFactCollect(t *testing.T, fct FactCollectTest) {
	t.Helper()
	assert := assert.New(t)
	// Hide logging output.
	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	fact.Facts = map[string]fact.Facter{}

	// Load input plugin.
	if fct.TestInput.Data != nil {
		testP := testdata.TestFacter{
			Name:                "test-input",
			TestInputDataFormat: fct.TestInput.DataFormat,
			TestInputData:       fct.TestInput.Data,
		}
		testP.Collect()
		fact.Facts["test-input"] = &testP
	}

	err := fct.Facter.ValidateInput()
	if fct.ExpectedInputError != nil {
		assert.Error(err, fct.ExpectedInputError)
		return
	} else {
		assert.Empty(err)
	}

	// Load additional inputs.
	if len(fct.TestAdditionalInputs) > 0 {
		for name, testInput := range fct.TestAdditionalInputs {
			testP := testdata.TestFacter{
				Name:                name,
				TestInputDataFormat: testInput.DataFormat,
				TestInputData:       testInput.Data,
			}
			testP.Collect()
			fact.Facts[name] = &testP
		}

		errs := fct.Facter.LoadAdditionalInputs()
		if len(fct.ExpectedAdditionalInputsErrs) > 0 {
			assert.ElementsMatch(fct.ExpectedAdditionalInputsErrs, errs)
			return
		} else {
			assert.Empty(errs)
		}
	}

	// Collect data.
	fct.Facter.Collect()
	assert.ElementsMatch(fct.ExpectedErrors, fct.Facter.GetErrors())
	assert.Equal(fct.ExpectedData, fct.Facter.GetData())
}
