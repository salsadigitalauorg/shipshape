package docker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/docker"
	"github.com/salsadigitalauorg/shipshape/pkg/fact/testdata"
)

func TestImagesCollect(t *testing.T) {
	assert := assert.New(t)

	type testInput struct {
		dataFormat data.DataFormat
		data       any
	}

	tests := []struct {
		name                 string
		testInput            testInput
		testAdditionalInputs map[string]testInput
		expectedInputError   error
		expectedErrors       []error
		expectedData         interface{}
	}{
		{
			name:               "noInput",
			expectedInputError: &fact.ErrSupportRequired{SupportType: "input"},
		},
		{
			name:           "inputFormatUnsupported",
			testInput:      testInput{dataFormat: data.FormatRaw, data: []byte("foo")},
			expectedErrors: []error{&fact.ErrSupportNone{SupportType: "input data format"}},
		},
		{
			name: "bogusData",
			testInput: testInput{
				dataFormat: data.FormatMapBytes,
				data:       map[string][]byte{"foo": []byte("bar")}},
			expectedData: map[string][]string{"foo": {}},
		},
		{
			name: "dockerfile/simple",
			testInput: testInput{
				dataFormat: data.FormatMapBytes,
				data:       map[string][]byte{"Dockerfile": []byte("FROM scratch\n")},
			},
			expectedData: map[string][]string{"Dockerfile": {"scratch:latest"}},
		},
		{
			name: "dockerfile/withArgs",
			testInput: testInput{
				dataFormat: data.FormatMapBytes,
				data: map[string][]byte{"php": []byte(`ARG CLI_IMAGE
ARG PHP_IMAGE_VERSION=8.3

FROM ${CLI_IMAGE} as cli
FROM php:${PHP_IMAGE_VERSION}
`)}},
			expectedData: map[string][]string{"php": {":latest", "php:8.3"}},
		},
		{
			name: "dockerfile/withArgsWithArgsInput",
			testInput: testInput{
				dataFormat: data.FormatMapBytes,
				data: map[string][]byte{"php": []byte(`ARG CLI_IMAGE
ARG PHP_IMAGE_VERSION=8.3

FROM ${CLI_IMAGE} as cli
FROM php:${PHP_IMAGE_VERSION}
`)},
			},
			testAdditionalInputs: map[string]testInput{
				"args-input": {
					dataFormat: data.FormatMapNestedString,
					data:       map[string]map[string]string{"php": {"CLI_IMAGE": "myapp"}},
				},
			},
			expectedData: map[string][]string{"php": {"myapp:latest", "php:8.3"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Images{Name: "base-images", InputName: "test-input"}
			fact.Facts = map[string]fact.Facter{}

			// Load input plugin.
			if tt.testInput.dataFormat != "" && tt.testInput.data != nil {
				testP := testdata.TestFacter{
					Name:                "test-input",
					TestInputDataFormat: tt.testInput.dataFormat,
					TestInputData:       tt.testInput.data,
				}
				testP.Collect()
				fact.Facts["test-input"] = &testP
			}

			err := p.ValidateInput()
			if tt.expectedInputError != nil {
				assert.Error(err, tt.expectedInputError)
				return
			} else {
				assert.Empty(err)
			}

			// Load additional inputs.
			if len(tt.testAdditionalInputs) > 0 {
				p.ArgsFrom = "args-input"
				p.AdditionalInputNames = []string{}
				for name, testInput := range tt.testAdditionalInputs {
					p.AdditionalInputNames = append(p.AdditionalInputNames, name)
					testP := testdata.TestFacter{
						Name:                name,
						TestInputDataFormat: testInput.dataFormat,
						TestInputData:       testInput.data,
					}
					testP.Collect()
					fact.Facts[name] = &testP
				}

				err := p.LoadAdditionalInputs()
				assert.Empty(err)
			}

			// Collect data.
			p.Collect()
			assert.ElementsMatch(tt.expectedErrors, p.GetErrors())
			assert.Equal(tt.expectedData, p.GetData())
		})
	}
}
