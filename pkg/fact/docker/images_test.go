package docker_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/docker"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

func TestImagesCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name: "noInput",
			FactFn: func() fact.Facter {
				f := NewImages("base-images")
				f.SetInputName("test-input")
				return f
			},
			ExpectedInputError: &plugin.ErrSupportRequired{SupportType: "input"},
		},
		{
			Name: "inputFormatUnsupported",
			FactFn: func() fact.Facter {
				f := NewImages("base-images")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{DataFormat: data.FormatRaw, Data: []byte("foo")},
			ExpectedInputError: &plugin.ErrSupportNone{
				Plugin:        "base-images",
				SupportType:   "input data format",
				SupportPlugin: "raw"},
		},
		{
			Name: "bogusData",
			FactFn: func() fact.Facter {
				f := NewImages("base-images")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"foo": []byte("bar")}},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData:   map[string][]string{"foo": {}},
		},
		{
			Name: "dockerfile/simple",
			FactFn: func() fact.Facter {
				f := NewImages("base-images")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"Dockerfile": []byte("FROM scratch\n")},
			},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData:   map[string][]string{"Dockerfile": {"scratch:latest"}},
		},
		{
			Name: "dockerfile/withArgs",
			FactFn: func() fact.Facter {
				f := NewImages("base-images")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{"php": []byte(`ARG CLI_IMAGE
ARG PHP_IMAGE_VERSION=8.3

FROM ${CLI_IMAGE} as cli
FROM php:${PHP_IMAGE_VERSION}
`)}},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData:   map[string][]string{"php": {":latest", "php:8.3"}},
		},
		{
			Name: "dockerfile/withArgs/WithAdditionalInput/NoDataFormat",
			FactFn: func() fact.Facter {
				f := NewImages("base-images")
				f.SetInputName("test-input")
				f.ArgsFrom = "args-input"
				f.AdditionalInputNames = []string{"args-input"}
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{"php": []byte(`ARG CLI_IMAGE
ARG PHP_IMAGE_VERSION=8.3

FROM ${CLI_IMAGE} as cli
FROM php:${PHP_IMAGE_VERSION}
`)},
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"args-input": {
					Data: map[string]map[string]string{"php": {"CLI_IMAGE": "myapp"}},
				},
			},
			ExpectedAdditionalInputsErrs: []error{&plugin.ErrSupportRequired{
				Plugin:      "testdata:testfacter",
				SupportType: "inputFormat"}},
		},
		{
			Name: "dockerfile/withArgs/WithAdditionalInput",
			FactFn: func() fact.Facter {
				f := NewImages("base-images")
				f.SetInputName("test-input")
				f.ArgsFrom = "args-input"
				f.AdditionalInputNames = []string{"args-input"}
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{"php": []byte(`ARG CLI_IMAGE
ARG PHP_IMAGE_VERSION=8.3

FROM ${CLI_IMAGE} as cli
FROM php:${PHP_IMAGE_VERSION}
`)},
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"args-input": {
					DataFormat: data.FormatMapNestedString,
					Data:       map[string]map[string]string{"php": {"CLI_IMAGE": "myapp"}},
				},
			},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData:   map[string][]string{"php": {"myapp:latest", "php:8.3"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}
