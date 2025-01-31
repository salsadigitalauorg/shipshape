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
)

func TestKeyInit(t *testing.T) {
	assert := assert.New(t)

	// Test that the yaml:key plugin is registered.
	factPlugin := fact.Registry["yaml:key"]("testKeyYaml")
	assert.NotNil(factPlugin)
	keyFacter, ok := factPlugin.(*Key)
	assert.True(ok)
	assert.Equal("testKeyYaml", keyFacter.Name)
}

func TestKeyPluginName(t *testing.T) {
	key := Key{Name: "testKeyYaml"}
	assert.Equal(t, "yaml:key", key.PluginName())
}

func TestKeySupportedConnections(t *testing.T) {
	key := Key{Name: "testKeyYaml"}
	supportLevel, connections := key.SupportedConnections()
	assert.Equal(t, fact.SupportNone, supportLevel)
	assert.Empty(t, connections)
}

func TestKeySupportedInputs(t *testing.T) {
	key := Key{Name: "testKeyYaml"}
	supportLevel, inputs := key.SupportedInputs()
	assert.Equal(t, fact.SupportRequired, supportLevel)
	assert.ElementsMatch(t, []string{"file:read", "file:lookup", "yaml:key"}, inputs)
}

func TestKeyCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name:               "noInput",
			Facter:             &Key{Name: "base-images"},
			ExpectedInputError: &fact.ErrSupportRequired{SupportType: "input"},
		},
		{
			Name:               "noInput/nameProvided",
			Facter:             &Key{Name: "base-images", InputName: "test-input"},
			ExpectedInputError: &fact.ErrSupportRequired{SupportType: "input"},
		},
		{
			Name:               "inputFormat/Empty",
			Facter:             &Key{Name: "base-images", InputName: "test-input"},
			TestInput:          internal.FactInputTest{Data: []byte("")},
			ExpectedInputError: &fact.ErrSupportRequired{SupportType: "input data format"},
		},

		// Raw data format (data.FormatRaw) cases.
		{
			Name:   "inputFormat/Raw/NotFound",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "foo"},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("bar: baz")},
			ExpectedErrors: []error{errors.New("yaml path not found")},
		},
		{
			Name: "inputFormat/Raw/NotFound/Ignored",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "foo",
				IgnoreNotFound: true},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("bar: baz")},
			ExpectedFormat: data.FormatNil,
			ExpectedData:   nil,
		},
		{
			Name:   "inputFormat/Raw",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "foo"},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("foo: bar")},
			ExpectedFormat: data.FormatString,
			ExpectedData:   "bar",
		},
		{
			Name: "inputFormat/Raw/NodesOnly",
			Facter: &Key{Name: "base-images", InputName: "test-input",
				Path: "foo", NodesOnly: true},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("foo: bar")},
			ExpectedFormat: FormatYamlNodes,
			ExpectedData: []*yaml.Node([]*yaml.Node{{
				Kind: yaml.ScalarNode, Tag: "!!str", Value: "bar", Line: 1, Column: 6,
			}}),
		},
		{
			Name: "inputFormat/Raw/KeysOnly",
			Facter: &Key{Name: "base-images", InputName: "test-input",
				Path: "foo", KeysOnly: true},
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
			Name:   "inputFormat/MapBytes/NotFound",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "foo"},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"file1": []byte("bar: baz")},
			},
			ExpectedErrors: []error{errors.New("yaml path not found")},
		},
		{
			Name: "inputFormat/MapBytes/NotFound/Ignored",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "foo",
				IgnoreNotFound: true},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"file1": []byte("bar: baz")},
			},
			ExpectedFormat: data.FormatNil,
			ExpectedData:   nil,
		},
		{
			Name:   "inputFormat/MapBytes/scalar",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "foo"},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"file1": []byte("foo: bar")},
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]any{"file1": "bar"},
		},
		{
			Name:   "inputFormat/MapBytes/list",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "foo"},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatMapBytes,
				Data:       map[string][]byte{"file1": []byte("foo: [bar, baz]")},
			},
			ExpectedFormat: data.FormatMapListString,
			ExpectedData:   map[string]any{"file1": []string{"bar", "baz"}},
		},
		{
			Name:   "inputFormat/MapBytes/mapOfString",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "foo"},
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
			Name:   "inputFormat/MapBytes/mapOfList",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "foo"},
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
			Name:   "inputFormat/YamlNodes/scalar",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "baz"},
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
			Name:   "inputFormat/YamlNodes/list",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "baz"},
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
			Name:   "inputFormat/YamlNodes/mapOfString",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "baz"},
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
			Name:   "inputFormat/YamlNodes/mapOfList",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "baz"},
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
			Name:   "inputFormat/MapYamlNodes/scalar",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "baz"},
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
			Name:   "inputFormat/MapYamlNodes/list",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "baz"},
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
			Name:   "inputFormat/MapYamlNodes/map",
			Facter: &Key{Name: "base-images", InputName: "test-input", Path: "baz"},
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
