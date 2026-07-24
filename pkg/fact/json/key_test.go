package json_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/json"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

func TestKeyInit(t *testing.T) {
	assert := assert.New(t)

	factPlugin := fact.Manager().GetFactories()["json:key"]("testKeyJson")
	assert.NotNil(factPlugin)
	keyFacter, ok := factPlugin.(*Key)
	assert.True(ok)
	assert.Equal("testKeyJson", keyFacter.GetId())
}

func TestKeyPluginName(t *testing.T) {
	key := New("testKeyJson")
	assert.Equal(t, "json:key", key.GetName())
}

func TestKeySupportedInputFormats(t *testing.T) {
	key := New("testKeyJson")
	supportLevel, inputFormats := key.SupportedInputFormats()
	assert.Equal(t, plugin.SupportRequired, supportLevel)
	assert.ElementsMatch(t, []data.DataFormat{
		data.FormatRaw,
		data.FormatMapBytes,
	}, inputFormats)
}

func TestKeyCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name:               "noInput",
			Facter:             New("json-values"),
			ExpectedInputError: &plugin.ErrSupportRequired{SupportType: "input"},
		},
		{
			Name: "noExpression",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"foo":"bar"}`)},
			ExpectedErrors: []error{errors.New("json:key requires an 'expression'")},
		},

		// Raw data format (data.FormatRaw) cases.
		{
			Name: "inputFormat/Raw/scalarString",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"foo":"bar"}`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"bar"},
		},
		{
			Name: "inputFormat/Raw/scalarNumber",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.count"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"count":42}`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"42"},
		},
		{
			Name: "inputFormat/Raw/scalarBool",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.private"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"private":true}`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"true"},
		},
		{
			Name: "inputFormat/Raw/arrayIndex",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.items[0]"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"items":["a","b","c"]}`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"a"},
		},
		{
			Name: "inputFormat/Raw/wildcardValues",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.scripts.*"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"scripts":{"build":"vite build","test":"vitest run"}}`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"vite build", "vitest run"},
		},
		{
			Name: "inputFormat/Raw/invalidJSON",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{not json`)},
			ExpectedErrors: []error{errors.New("invalid JSON input")},
		},

		// Map of Raw data (data.FormatMapBytes) format cases.
		{
			Name: "inputFormat/MapBytes/scalar",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"file1": []byte(`{"foo":"bar"}`)},
			},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData:   map[string][]string{"file1": {"bar"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}
