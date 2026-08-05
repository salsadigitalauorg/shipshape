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
			ExpectedFormat: data.FormatString,
			ExpectedData:   "bar",
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
			ExpectedFormat: data.FormatString,
			ExpectedData:   "42",
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
			ExpectedFormat: data.FormatString,
			ExpectedData:   "true",
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
			ExpectedFormat: data.FormatString,
			ExpectedData:   "a",
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
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{"file1": "bar"},
		},
		{
			Name: "inputFormat/MapBytes/list",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.items[*]"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"file1": []byte(`{"items":["c","a"]}`)},
			},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData:   map[string][]string{"file1": {"c", "a"}},
		},

		// Multi-file cases where every file's match shares the same shape: the
		// map format is promoted from that shared shape, and no file's result
		// depends on which order the files happen to be visited in.
		{
			Name: "inputFormat/MapBytes/multiFile/uniformScalar",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.name"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"a.json": []byte(`{"name":"alpha"}`),
					"b.json": []byte(`{"name":"beta"}`),
					"c.json": []byte(`{"name":"gamma"}`),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"a.json": "alpha", "b.json": "beta", "c.json": "gamma"},
		},
		{
			Name: "inputFormat/MapBytes/multiFile/uniformObject",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.a"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"x.json": []byte(`{"a":{"k":"v"}}`),
					"y.json": []byte(`{"a":{"k2":"v2"}}`),
				},
			},
			ExpectedFormat: data.FormatMapNestedString,
			ExpectedData: map[string]map[string]string{
				"x.json": {"k": "v"}, "y.json": {"k2": "v2"}},
		},

		// Multi-file case where files match with DIFFERING shapes (one scalar,
		// one object). This is a regression test for a bug where the combined
		// map format was decided by whichever file Go's randomised map
		// iteration visited first, silently discarding every file that did not
		// match that shape - so the same input could produce 1 or 2 result
		// entries depending on run. Every file's finding must always be
		// present, and the format must be the same on every run.
		{
			Name: "inputFormat/MapBytes/multiFile/mixedShapesNoDataLoss",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.a"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"x.json": []byte(`{"a":{"k":"v"}}`),
					"y.json": []byte(`{"a":"scalarValue"}`),
				},
			},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData: map[string][]string{
				"x.json": {"k=v"}, "y.json": {"scalarValue"}},
		},

		// A file where the expression matches nothing must not affect the
		// shape of the files that do match.
		{
			Name: "inputFormat/MapBytes/multiFile/oneFileNoMatch",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.name"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"a.json":       []byte(`{"name":"alpha"}`),
					"noMatch.json": []byte(`{"other":"value"}`),
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{"a.json": "alpha"},
		},

		// If every file's expression matches nothing, the fact resolves to nil
		// rather than an empty map.
		{
			Name: "inputFormat/MapBytes/multiFile/allFilesNoMatch",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.missing"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{
					"a.json": []byte(`{"name":"alpha"}`),
					"b.json": []byte(`{"name":"beta"}`),
				},
			},
			ExpectedFormat: data.FormatNil,
		},

		// Malformed expressions. With fail-early false (the default) the error
		// is recorded rather than aborting the run.
		{
			Name: "invalidExpression",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.items[?(("
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"items":["a"]}`)},
			ExpectedErrors: []error{errors.New("invalid JSONPath expression")},
		},

		// No match emits FormatNil rather than an error, so that analysers such
		// as not-empty can act on the absence of a value.
		{
			Name: "inputFormat/Raw/noMatch",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.missing"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"foo":"bar"}`)},
			ExpectedFormat: data.FormatNil,
		},

		// Array order must be preserved, not sorted.
		{
			Name: "inputFormat/Raw/arrayOrderPreserved",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.items[*]"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"items":["c","a","b"]}`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"c", "a", "b"},
		},
		{
			Name: "inputFormat/Raw/matchedArrayOrderPreserved",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.items"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"items":["c","a","b"]}`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"c", "a", "b"},
		},

		// Object and nested-object shapes, mirroring yaml:key's formats.
		{
			Name: "inputFormat/Raw/object",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.scripts"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"scripts":{"build":"vite build","test":"vitest run"}}`)},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"build": "vite build", "test": "vitest run"},
		},
		{
			Name: "inputFormat/Raw/objectOfObjects",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.envs"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"envs":{"dev":{"url":"d.example"},"prod":{"url":"p.example"}}}`)},
			ExpectedFormat: data.FormatMapNestedString,
			ExpectedData: map[string]map[string]string{
				"dev":  {"url": "d.example"},
				"prod": {"url": "p.example"},
			},
		},

		// A flat object containing a value that is itself an object (deeper
		// than objectOfObjects can represent as FormatMapNestedString) or an
		// array must not render Go's internal syntax (e.g. "map[k:v]", "[1 2]")
		// into the emitted value - doing so could leak nested sensitive values
		// (such as a credential) verbatim into breach output. This mirrors
		// yaml:key, which reads a YAML scalar node's Value and gets "" for a
		// non-scalar node.
		{
			Name: "inputFormat/Raw/objectWithNonFlatValueNoGoSyntaxLeak",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.config"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data: []byte(`{"config":` +
					`{"db":{"host":"localhost","password":"s3cr3t"},"name":"app"}}`)},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"db": "", "name": "app"},
		},
		{
			Name: "inputFormat/Raw/objectWithArrayValueNoGoSyntaxLeak",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.a"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"a":{"k":[1,2],"n":"v"}}`)},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{"k": "", "n": "v"},
		},
		{
			Name: "inputFormat/Raw/listOfObjects",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.deps[*]"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"deps":[{"n":"x","v":"1.0"},{"n":"y","v":"2.0"}]}`)},
			ExpectedFormat: data.FormatListMapString,
			ExpectedData: []map[string]string{
				{"n": "x", "v": "1.0"},
				{"n": "y", "v": "2.0"},
			},
		},

		// RFC 9535 syntax that the previous (pre-RFC) library rejected.
		{
			Name: "rfc9535/filterWithoutParens",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.deps[?@.n=='x'].v"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"deps":[{"n":"x","v":"1.0"},{"n":"y","v":"2.0"}]}`)},
			ExpectedFormat: data.FormatString,
			ExpectedData:   "1.0",
		},
		{
			Name: "rfc9535/filterWithWhitespace",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.deps[?@.n == 'y'].v"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"deps":[{"n":"x","v":"1.0"},{"n":"y","v":"2.0"}]}`)},
			ExpectedFormat: data.FormatString,
			ExpectedData:   "2.0",
		},
		{
			Name: "rfc9535/matchFunction",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.scripts[?match(@,'vite.*')]"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"scripts":{"build":"vite build","lint":"eslint ."}}`)},
			ExpectedFormat: data.FormatString,
			ExpectedData:   "vite build",
		},
		{
			Name: "rfc9535/searchFunction",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.scripts[?search(@,'eslint')]"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"scripts":{"build":"vite build","lint":"eslint ."}}`)},
			ExpectedFormat: data.FormatString,
			ExpectedData:   "eslint .",
		},
		{
			Name: "rfc9535/union",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$.items[0,2]"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`{"items":["a","b","c"]}`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"a", "c"},
		},
		{
			Name: "rfc9535/descendant",
			FactFn: func() fact.Facter {
				f := New("json-values")
				f.SetInputName("test-input")
				f.Expression = "$..n"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte(`{"deps":[{"n":"x"},{"n":"y"}]}`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"x", "y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}

// TestKeyValidateExpression covers expression parsing in isolation, including
// the malformed case that must return an error rather than panic.
func TestKeyValidateExpression(t *testing.T) {
	assert := assert.New(t)

	k := New("json-values")
	assert.ErrorIs(k.ValidateExpression(), ErrNoExpression)

	k.Expression = "$.items[?(("
	assert.ErrorIs(k.ValidateExpression(), ErrInvalidExpression)

	k = New("json-values")
	k.Expression = "$.deps[?@.n=='x'].v"
	assert.NoError(k.ValidateExpression())
	// Repeated calls reuse the already-parsed path.
	assert.NoError(k.ValidateExpression())
}
