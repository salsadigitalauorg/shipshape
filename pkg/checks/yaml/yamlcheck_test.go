package yaml_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
)

var cFalse = false
var cTrue = true

func TestYamlCheckMerge(t *testing.T) {
	assert := assert.New(t)

	c := YamlCheck{
		Path:           "path1",
		File:           "file1.yml",
		Pattern:        "pattern1",
		ExcludePattern: "excludePattern1",
		IgnoreMissing:  &cFalse,
	}
	c.Merge(&YamlCheck{
		Path:  "path2",
		Files: []string{"slcFile1.yml", "slcFile2.yml"},
	})
	assert.EqualValues(YamlCheck{
		Path:           "path2",
		File:           "file1.yml",
		Files:          []string{"slcFile1.yml", "slcFile2.yml"},
		Pattern:        "pattern1",
		ExcludePattern: "excludePattern1",
		IgnoreMissing:  &cFalse,
	}, c)
}

func TestYamlCheckFetchData(t *testing.T) {
	tt := []internal.FetchDataTest{
		{
			Name: "noFile",
			Check: &YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
			},
			ExpectBreaches: []result.Breach{
				&result.ValueBreach{
					BreachType: "value",
					CheckType:  "yaml",
					Severity:   "normal",
					ValueLabel: "- no file",
					Value:      "no file provided",
				},
			},
			ExpectStatusFail: true,
		},

		{
			Name: "nonExistentFile",
			Check: &YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				File: "non-existent.yml",
			},
			ExpectBreaches: []result.Breach{
				&result.ValueBreach{
					BreachType: "value",
					CheckType:  "yaml",
					Severity:   "normal",
					ValueLabel: "error reading file: testdata/non-existent.yml",
					Value:      "open testdata/non-existent.yml: no such file or directory",
				},
			},
			ExpectStatusFail: true,
		},

		{
			Name: "nonExistentFileIgnoreMissing",
			Check: &YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				File:          "non-existent.yml",
				IgnoreMissing: &cTrue,
			},
			ExpectPasses:     []string{"File testdata/non-existent.yml does not exist"},
			ExpectStatusPass: true,
		},

		{
			Name: "singleFile",
			Check: &YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				File: "update.settings.yml",
			},
			ExpectDataMap: map[string][]byte{
				"update.settings.yml": []byte(
					`check:
  interval_days: 7
notification:
  emails:
    - admin@example.com
`),
			},
		},

		{
			Name: "badFilePattern",
			Check: &YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern: "*.bar.yml",
				Path:    "",
			},
			ExpectBreaches: []result.Breach{
				&result.ValueBreach{
					BreachType: "value",
					CheckType:  "yaml",
					Severity:   "normal",
					ValueLabel: "error finding files in path: testdata",
					Value:      "error parsing regexp: missing argument to repetition operator: `*`",
				},
			},
			ExpectStatusFail: true,
		},

		{
			Name: "filePatternNoMatchingFile",
			Check: &YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern: "bla.*.yml",
			},
			ExpectBreaches: []result.Breach{
				&result.ValueBreach{
					BreachType: "value",
					CheckType:  "yaml",
					Severity:   "normal",
					ValueLabel: "- no file",
					Value:      "no matching yaml files found",
				},
			},
			ExpectStatusFail: true,
		},

		{
			Name: "filePatternNoMatchingFileIgnoreMissing",
			Check: &YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern:       "bla.*.yml",
				IgnoreMissing: &cTrue,
			},
			ExpectPasses:     []string{"no matching config files found"},
			ExpectStatusPass: true,
		},

		{
			Name: "correctSingleFilePatternAndValue",
			Check: &YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern: "foo.bar.yml",
				Path:    "dir/subdir",
			},
			ExpectDataMap: map[string][]byte{
				"testdata/dir/subdir/foo.bar.yml": []byte(
					`check:
  interval_days: 7
`),
			},
		},

		{
			Name: "recursiveFileLookup",
			Check: &YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern: ".*.bar.yml",
			},
			ExpectDataMap: map[string][]byte{
				"testdata/dir/foo.bar.yml": []byte(
					`check:
  interval_days: 7
`),
				"testdata/dir/subdir/foo.bar.yml": []byte(
					`check:
  interval_days: 7
`),
				"testdata/dir/subdir/zoom.bar.yml": []byte(
					`check:
  interval_days: 5
`),
				"testdata/dir/zoom.bar.yml": []byte(
					`check:
  interval_days: 5
`),
				"testdata/foo.bar.yml": []byte(
					`check:
  interval_days: 7
`),
				"testdata/zoom.bar.yml": []byte(
					`check:
  interval_days: 5
`),
			},
		},
	}

	config.ProjectDir = "testdata"
	for _, tc := range tt {
		t.Run(tc.Name, func(innerT *testing.T) {
			tc.Check.Init(Yaml)
			internal.TestFetchData(innerT, tc)
		})
	}
}
