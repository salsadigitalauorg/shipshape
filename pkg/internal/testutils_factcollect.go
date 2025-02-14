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
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
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

	fact.Facts = map[string]fact.Facter{}

	if fct.FactFn != nil {
		fct.Facter = fct.FactFn()
	}

	// Load input plugin.
	if fct.TestInput.DataFn != nil {
		fct.TestInput.Data = fct.TestInput.DataFn()
	}

	if fct.TestInput.Data != nil {
		testP := testdata.TestFacter{
			BaseFact: fact.BaseFact{
				BasePlugin: plugin.BasePlugin{
					Id: "test-input",
				},
			},
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
				BaseFact: fact.BaseFact{
					BasePlugin: plugin.BasePlugin{
						Id: name,
					},
				},
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
