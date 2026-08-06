package yaml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

func TestLintInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the yaml:lint plugin is registered.
	factPlugin := fact.Manager().GetFactories()["yaml:lint"]("testLintYaml")
	assert.NotNil(factPlugin)
	lintFacter, ok := factPlugin.(*Lint)
	assert.True(ok)
	assert.Equal("testLintYaml", lintFacter.GetId())
}

func TestLintPluginName(t *testing.T) {
	lint := NewLint("testLintYaml")
	assert.Equal(t, "yaml:lint", lint.GetName())
}

func TestLintSupportedConnections(t *testing.T) {
	lint := NewLint("testLintYaml")
	supportLevel, connections := lint.SupportedConnections()
	assert.Equal(t, plugin.SupportNone, supportLevel)
	assert.Empty(t, connections)
}

func TestLintSupportedInputFormats(t *testing.T) {
	lint := NewLint("testLintYaml")
	supportLevel, inputFormats := lint.SupportedInputFormats()
	assert.Equal(t, plugin.SupportRequired, supportLevel)
	assert.ElementsMatch(t, []data.DataFormat{data.FormatMapBytes}, inputFormats)
}

// TestLintAllDocumentsDefault asserts the constructor opts in to
// multi-document validation, since that is the documented default and the
// behaviour that diverges from 0.x.
func TestLintAllDocumentsDefault(t *testing.T) {
	lint := NewLint("testLintYaml")
	assert.NotNil(t, lint.AllDocuments)
	assert.True(t, *lint.AllDocuments)
}

func TestLintCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name:               "noInput",
			Facter:             NewLint("invalid-yaml"),
			ExpectedInputError: &plugin.ErrSupportRequired{SupportType: "input"},
		},
		{
			Name: "noInput/nameProvided",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			ExpectedInputError: &plugin.ErrSupportRequired{SupportType: "input"},
		},
		// An unsupported format is rejected by ValidateInput before Collect
		// runs (pkg/fact/base.go:173-184), so the plugin's own default branch
		// is defensive only. Asserting the input error here documents where
		// the rejection actually happens.
		{
			Name: "inputFormat/unsupported",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("foo: bar")},
			ExpectedInputError: &plugin.ErrSupportNone{
				Plugin:        "yaml:lint",
				SupportType:   "inputFormat",
				SupportPlugin: string(data.FormatRaw)},
		},

		// A valid tree emits an empty map rather than nil, so not:empty sees a
		// supported format and simply finds nothing to breach on.
		{
			Name: "allValid",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"a.yml": []byte("foo: bar\n"),
					"b.yml": []byte("baz:\n  - one\n  - two\n"),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{},
		},
		{
			Name: "emptyInput",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{},
		},
		// An empty file is valid YAML (decodes to nil), matching 0.x.
		{
			Name: "emptyFileIsValid",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"empty.yml": []byte("")},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{},
		},

		// A parse failure must be data, never a collection error - otherwise
		// the run would abort before reaching the analyse stage, which is the
		// gap this plugin exists to close.
		{
			Name: "syntaxError",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"good.yml": []byte("foo: bar\n"),
					"bad.yml":  []byte("foo: [unclosed\n"),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"bad.yml": "yaml: line 1: did not find expected ',' or ']'",
			},
		},
		{
			Name: "mappingValuesNotAllowed",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"bad.yml": []byte("a: b: c\n")},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"bad.yml": "yaml: mapping values are not allowed in this context",
			},
		},
		// Duplicate keys surface as *yaml.TypeError, which carries the 0.x
		// "cannot decode yaml: " prefix rather than the bare parser message.
		{
			Name: "duplicateKeysAreTypeError",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"dupe.yml": []byte("a: 1\na: 2\n")},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"dupe.yml": `cannot decode yaml: line 2: mapping key "a" already defined at line 1`,
			},
		},
		{
			Name: "multipleInvalidFiles",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"ok.yml":  []byte("foo: bar\n"),
					"one.yml": []byte("foo: [unclosed\n"),
					"two.yml": []byte("a: 1\na: 2\n"),
					"tab.yml": []byte("a:\n\tb: 1\n"),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"one.yml": "yaml: line 1: did not find expected ',' or ']'",
				"two.yml": `cannot decode yaml: line 2: mapping key "a" already defined at line 1`,
				"tab.yml": "yaml: line 2: found character that cannot start any token",
			},
		},

		// Multi-document handling: the error is in the SECOND document, so
		// only the decoder loop finds it. This is the deliberate divergence
		// from 0.x's single yaml.Unmarshal.
		//
		// The reported line (2) is one less than the file line (3): yaml.v3
		// reports where the unclosed '[' opened. Asserted verbatim so a
		// change in the library's positions is caught rather than masked.
		{
			Name: "multiDocument/lateErrorCaughtByDefault",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"multi.yml": []byte("a: 1\n---\nb: [unclosed\n"),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"multi.yml": "yaml: line 2: did not find expected ',' or ']'",
			},
		},
		{
			Name: "multiDocument/lateErrorMissedWhenDisabled",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				allDocs := false
				f.AllDocuments = &allDocs
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"multi.yml": []byte("a: 1\n---\nb: [unclosed\n"),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{},
		},
		{
			Name: "multiDocument/allValid",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"multi.yml": []byte("a: 1\n---\nb: 2\n---\nc: 3\n"),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{},
		},
		// An error in the FIRST document is found regardless of the
		// all-documents setting; the pair below asserts both paths agree.
		{
			Name: "multiDocument/firstDocError",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"multi.yml": []byte("a: [unclosed\n---\nb: 2\n"),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"multi.yml": "yaml: line 1: did not find expected ',' or ']'",
			},
		},
		// An error in the first document is found either way.
		{
			Name: "multiDocument/earlyErrorFoundWhenDisabled",
			FactFn: func() fact.Facter {
				f := NewLint("invalid-yaml")
				f.SetInputName("test-input")
				allDocs := false
				f.AllDocuments = &allDocs
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"multi.yml": []byte("a: [unclosed\n---\nb: 2\n"),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"multi.yml": "yaml: line 1: did not find expected ',' or ']'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}

// TestLintAllDocumentsNilDefaultsTrue covers a Lint built by the fact
// manager's yaml.Unmarshal path rather than NewLint: a config that omits
// 'all-documents' leaves the pointer nil, which must still validate every
// document.
func TestLintAllDocumentsNilDefaultsTrue(t *testing.T) {
	internal.TestFactCollect(t, internal.FactCollectTest{
		Name: "allDocumentsNil",
		FactFn: func() fact.Facter {
			f := NewLint("invalid-yaml")
			f.SetInputName("test-input")
			f.AllDocuments = nil
			return f
		},
		TestInput: internal.FactInputTest{
			DataFormat: data.FormatMapBytes,
			Data: map[string][]byte{
				"multi.yml": []byte("a: 1\n---\nb: [unclosed\n"),
			},
		},
		ExpectedFormat: data.FormatMapString,
		ExpectedData: map[string]string{
			"multi.yml": "yaml: line 2: did not find expected ',' or ']'",
		},
	})
}
