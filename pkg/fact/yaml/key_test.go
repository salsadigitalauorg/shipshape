package yaml_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

func TestKeyInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the yaml:key plugin is registered.
	factPlugin := fact.Registry["yaml:key"]("testKeyYaml")
	assert.NotNil(factPlugin)
	keyFacter, ok := factPlugin.(*Key)
	assert.True(ok)
	assert.Equal("testKeyYaml", keyFacter.GetId())
}

func TestKeyPluginName(t *testing.T) {
	key := New("testKeyYaml")
	assert.Equal(t, "yaml:key", key.GetName())
}

func TestKeySupportedConnections(t *testing.T) {
	key := New("testKeyYaml")
	supportLevel, connections := key.SupportedConnections()
	assert.Equal(t, plugin.SupportNone, supportLevel)
	assert.Empty(t, connections)
}

func TestKeySupportedInputs(t *testing.T) {
	key := New("testKeyYaml")
	supportLevel, inputs := key.SupportedInputs()
	assert.Equal(t, plugin.SupportRequired, supportLevel)
	assert.ElementsMatch(t, []string{
		"docker:command",
		"file:read",
		"file:lookup",
		"yaml:key"}, inputs)
}

func TestKeyCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name:               "noInput",
			Facter:             New("base-images"),
			ExpectedInputError: &plugin.ErrSupportRequired{SupportType: "input"},
		},
		{
			Name: "noInput/nameProvided",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				return f
			},
			ExpectedInputError: &plugin.ErrSupportRequired{SupportType: "input"},
		},
		{
			Name: "inputFormat/Empty",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				return f
			},
			TestInput:          internal.FactInputTest{Data: []byte("")},
			ExpectedInputError: &plugin.ErrSupportRequired{SupportType: "input data format"},
		},

		// Raw data format (data.FormatRaw) cases.
		{
			Name: "inputFormat/Raw/NotFound",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("bar: baz")},
			ExpectedErrors: []error{errors.New("yaml path not found")},
		},
		{
			Name: "inputFormat/Raw/NotFound/Ignored",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				f.IgnoreNotFound = true
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("bar: baz")},
			ExpectedFormat: data.FormatNil,
			ExpectedData:   nil,
		},
		{
			Name: "inputFormat/Raw",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("foo: bar")},
			ExpectedFormat: data.FormatString,
			ExpectedData:   "bar",
		},
		{
			Name: "inputFormat/Raw/NodesOnly",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				f.NodesOnly = true
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("foo: bar")},
			ExpectedFormat: FormatYamlNodes,
			ExpectedData: []*yaml.Node([]*yaml.Node{{
				Kind: yaml.ScalarNode, Tag: "!!str", Value: "bar", Line: 1, Column: 6,
			}}),
		},
		{
			Name: "inputFormat/Raw/KeysOnly",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				f.KeysOnly = true
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte(`foo:
  bar: baz
  zoo: bar
`)},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{"bar", "zoo"},
		},

		// Map of Raw data (data.FormatMapBytes) format cases.
		{
			Name: "inputFormat/MapBytes/NotFound",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"file1": []byte("bar: baz")},
			},
			ExpectedErrors: []error{errors.New("yaml path not found")},
		},
		{
			Name: "inputFormat/MapBytes/NotFound/Ignored",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				f.IgnoreNotFound = true
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"file1": []byte("bar: baz")},
			},
			ExpectedFormat: data.FormatNil,
			ExpectedData:   nil,
		},
		{
			Name: "inputFormat/MapBytes/scalar",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"file1": []byte("foo: bar")},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]any{"file1": "bar"},
		},
		{
			Name: "inputFormat/MapBytes/list",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"file1": []byte("foo: [bar, baz]")},
			},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData:   map[string]any{"file1": []string{"bar", "baz"}},
		},
		{
			Name: "inputFormat/MapBytes/mapOfString",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{"file1": []byte(`foo:
  bar: baz
  zoo: bar
`)},
			},
			ExpectedFormat: data.FormatMapNestedString,
			ExpectedData:   map[string]any{"file1": map[string]string{"bar": "baz", "zoo": "bar"}},
		},
		{
			Name: "inputFormat/MapBytes/mapOfList",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "foo"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data: map[string][]byte{"file1": []byte(`foo:
  bar: [baz, zoom]
  zoo: [bar, zap]
`)},
			},
			ExpectedFormat: data.FormatMapNestedString,
			ExpectedData:   map[string]any{"file1": map[string]string{"bar": "", "zoo": ""}},
		},

		// List of Yaml nodes (FormatYamlNodes) format cases.
		{
			Name: "inputFormat/YamlNodes/scalar",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "baz"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: FormatYamlNodes,
				DataFn: func() any {
					var node1 yaml.Node
					yaml.Unmarshal([]byte("foo: \n  baz: zoom"), &node1)
					var node2 yaml.Node
					yaml.Unmarshal([]byte("bar: \n  baz: zap"), &node2)
					return []*yaml.Node{node1.Content[0], node2.Content[0]}
				},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]any{"bar": "zap", "foo": "zoom"},
		},
		{
			Name: "inputFormat/YamlNodes/list",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "baz"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: FormatYamlNodes,
				DataFn: func() any {
					var node1 yaml.Node
					yaml.Unmarshal([]byte("foo: \n  baz: [zoom,zap]"), &node1)
					var node2 yaml.Node
					yaml.Unmarshal([]byte("bar: \n  baz: [whoop,pop]"), &node2)
					return []*yaml.Node{node1.Content[0], node2.Content[0]}
				},
			},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData: map[string]any{
				"foo": []string{"zoom", "zap"},
				"bar": []string{"whoop", "pop"}},
		},
		{
			Name: "inputFormat/YamlNodes/mapOfString",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "baz"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: FormatYamlNodes,
				DataFn: func() any {
					var node1 yaml.Node
					yaml.Unmarshal([]byte(`foo:
  baz:
    zoom: zap
`), &node1)
					var node2 yaml.Node
					yaml.Unmarshal([]byte(`bar:
  baz:
    whoop: pop
`), &node2)
					return []*yaml.Node{node1.Content[0], node2.Content[0]}
				},
			},
			ExpectedFormat: data.FormatMapNestedString,
			ExpectedData: map[string]any{
				"foo": map[string]string{"zoom": "zap"},
				"bar": map[string]string{"whoop": "pop"}},
		},
		{
			Name: "inputFormat/YamlNodes/mapOfList",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "baz"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: FormatYamlNodes,
				DataFn: func() any {
					var node1 yaml.Node
					yaml.Unmarshal([]byte(`foo:
  baz:
    zoom: [zap,zip]
`), &node1)
					var node2 yaml.Node
					yaml.Unmarshal([]byte(`bar:
  baz:
    whoop: [pop,pip]
`), &node2)
					return []*yaml.Node{node1.Content[0], node2.Content[0]}
				},
			},
			ExpectedFormat: data.FormatMapNestedString,
			ExpectedData: map[string]any{
				"foo": map[string]string{"zoom": ""},
				"bar": map[string]string{"whoop": ""}},
		},

		// Map of Yaml nodes (FormatMapYamlNodes) format cases.
		{
			Name: "inputFormat/MapYamlNodes/scalar",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "baz"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: FormatMapYamlNodes,
				DataFn: func() any {
					var node1 yaml.Node
					yaml.Unmarshal([]byte("foo: \n  baz: zoom"), &node1)
					var node2 yaml.Node
					yaml.Unmarshal([]byte("bar: \n  baz: zap"), &node2)
					var node3 yaml.Node
					yaml.Unmarshal([]byte("zoom: \n  baz: paf"), &node3)
					var node4 yaml.Node
					yaml.Unmarshal([]byte("whoop: \n  baz: blo"), &node4)
					return map[string][]*yaml.Node{
						"key1": {node1.Content[0], node2.Content[0]},
						"key2": {node3.Content[0], node4.Content[0]},
					}
				},
			},
			ExpectedFormat: data.FormatMapNestedString,
			ExpectedData: map[string]map[string]string{
				"key1": {"bar": "zap", "foo": "zoom"},
				"key2": {"zoom": "paf", "whoop": "blo"}},
		},
		{
			Name: "inputFormat/MapYamlNodes/list",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "baz"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: FormatMapYamlNodes,
				DataFn: func() any {
					var node1 yaml.Node
					yaml.Unmarshal([]byte("foo: \n  baz: [zoom,aww]"), &node1)
					var node2 yaml.Node
					yaml.Unmarshal([]byte("bar: \n  baz: [zap,pif]"), &node2)
					var node3 yaml.Node
					yaml.Unmarshal([]byte("zoom: \n  baz: [paf,pop]"), &node3)
					var node4 yaml.Node
					yaml.Unmarshal([]byte("whoop: \n  baz: [blo,bro]"), &node4)
					return map[string][]*yaml.Node{
						"key1": {node1.Content[0], node2.Content[0]},
						"key2": {node3.Content[0], node4.Content[0]},
					}
				},
			},
			ExpectedErrors: []error{
				errors.New("unsupported format for nested lookup")},
		},
		{
			Name: "inputFormat/MapYamlNodes/map",
			FactFn: func() fact.Facter {
				f := New("base-images")
				f.SetInputName("test-input")
				f.Path = "baz"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: FormatMapYamlNodes,
				DataFn: func() any {
					var node1 yaml.Node
					yaml.Unmarshal([]byte("foo:\n  baz:\n    zoom: aww"), &node1)
					var node2 yaml.Node
					yaml.Unmarshal([]byte("bar:\n  baz:\n    zap: pif"), &node2)
					var node3 yaml.Node
					yaml.Unmarshal([]byte("zoom:\n  baz:\n    paf: pop"), &node3)
					var node4 yaml.Node
					yaml.Unmarshal([]byte("whoop:\n  baz:\n    blo: bro"), &node4)
					return map[string][]*yaml.Node{
						"key1": {node1.Content[0], node2.Content[0]},
						"key2": {node3.Content[0], node4.Content[0]},
					}
				},
			},
			ExpectedErrors: []error{
				errors.New("unsupported format for nested lookup")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}
