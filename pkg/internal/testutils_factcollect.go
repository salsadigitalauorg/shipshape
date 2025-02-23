package internal

import (
	"io"
	"reflect"
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
	DataFn     func() any
}

type FactCollectTest struct {
	Name   string
	Facter fact.Facter
	FactFn func() fact.Facter

	TestInput          FactInputTest
	ExpectedInputError error

	TestAdditionalInputs         map[string]FactInputTest
	ExpectedAdditionalInputsErrs []error

	ExpectedErrors []error
	ExpectedFormat data.DataFormat
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

	fact.Manager().ResetPlugins()

	if fct.FactFn != nil {
		fct.Facter = fct.FactFn()
	}

	// Load input plugin.
	if fct.TestInput.DataFn != nil {
		fct.TestInput.Data = fct.TestInput.DataFn()
	}

	if fct.TestInput.Data != nil {
		p, err := fact.Manager().GetPlugin("testdata:testfacter", "test-input")
		if err != nil {
			t.Fatalf("failed to get test input plugin: %s", err)
		}

		testP := p.(*testdata.TestFacter)
		testP.TestInputDataFormat = fct.TestInput.DataFormat
		testP.TestInputData = fct.TestInput.Data
		testP.Collect()
	}

	err := fact.ValidateInput(fct.Facter)
	if fct.ExpectedInputError != nil {
		assert.Error(err, fct.ExpectedInputError)
		return
	} else {
		assert.Empty(err)
	}

	// Load additional inputs.
	if len(fct.TestAdditionalInputs) > 0 {
		for name, testInput := range fct.TestAdditionalInputs {
			p, err := fact.Manager().GetPlugin("testdata:testfacter", name)
			if err != nil {
				t.Fatalf("failed to get test input plugin: %s", err)
			}

			testP := p.(*testdata.TestFacter)
			testP.TestInputDataFormat = testInput.DataFormat
			testP.TestInputData = testInput.Data
			testP.Collect()
		}

		errs := fact.LoadAdditionalInputs(fct.Facter)
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
	assert.Equal(fct.ExpectedFormat, fct.Facter.GetFormat())

	if fct.ExpectedData == nil {
		return
	}
	kindData := reflect.TypeOf(fct.ExpectedData).Kind()
	if kindData == reflect.Slice {
		assert.ElementsMatch(fct.ExpectedData, fct.Facter.GetData())
	} else {
		assert.Equal(fct.ExpectedData, fct.Facter.GetData())
	}
}
