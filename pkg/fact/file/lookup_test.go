package file_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/file"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/stretchr/testify/assert"
)

func TestFileLookupInit(t *testing.T) {
	assert := assert.New(t)

	factPlugin := fact.Manager().GetFactories()["file:lookup"]("testFileLookup")
	assert.NotNil(factPlugin)
	fileLookupFacter, ok := factPlugin.(*Lookup)
	assert.True(ok)
	assert.Equal("testFileLookup", fileLookupFacter.GetId())
}

func TestFileLookupPluginName(t *testing.T) {
	fileLookup := NewLookup("testFileLookup")
	assert.Equal(t, "file:lookup", fileLookup.GetName())
}

func TestFileLookupCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name: "testLookup",
			FactFn: func() fact.Facter {
				f := NewLookup("TestLookup")
				f.Path = "testdata/private"
				f.Pattern = ".*\\.(sql|php|sh|py|bz2|gz|tar|tgz|zip)?$"
				return f
			},
			ExpectedFormat: data.FormatListString,
			ExpectedData: []string{
				"testdata/private/dump.sql",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testLookupSkipDirs0",
			FactFn: func() fact.Facter {
				f := NewLookup("testLookupSkipDirs0")
				f.Path = "testdata/nested"
				f.Pattern = ".*\\.sql"
				f.SkipDirs = []string{}
				return f
			},
			ExpectedFormat: data.FormatListString,
			ExpectedData: []string{
				"testdata/nested/nested-01/nested-02/skipped/db.sql",
				"testdata/nested/skipped/dump.sql",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testLookupSkipDirs1",
			FactFn: func() fact.Facter {
				f := NewLookup("testLookupSkipDirs1")
				f.Path = "testdata/nested"
				f.Pattern = ".*\\.sql"
				f.SkipDirs = []string{
					"skipped",
				}
				return f
			},
			ExpectedFormat: data.FormatListString,
			ExpectedData: []string{
				"testdata/nested/nested-01/nested-02/skipped/db.sql",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testLookupSkipDirs2",
			FactFn: func() fact.Facter {
				f := NewLookup("testLookupSkipDirs2")
				f.Path = "testdata/nested"
				f.Pattern = ".*\\.sql"
				f.SkipDirs = []string{
					"skipped",
					"nested-01/nested-02/skipped",
				}
				return f
			},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{},
			ExpectedErrors: []error{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}
