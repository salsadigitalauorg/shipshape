package yaml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
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
			ExpectBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckType:  "yaml",
					Severity:   "normal",
					ValueLabel: "- no file",
					Value:      "no file provided",
				},
			},
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
			ExpectBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckType:  "yaml",
					Severity:   "normal",
					ValueLabel: "error reading file: testdata/non-existent.yml",
					Value:      "open testdata/non-existent.yml: no such file or directory",
				},
			},
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
			ExpectPasses: []string{"File testdata/non-existent.yml does not exist"},
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
			ExpectBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckType:  "yaml",
					Severity:   "normal",
					ValueLabel: "error finding files in path: testdata",
					Value:      "error parsing regexp: missing argument to repetition operator: `*`",
				},
			},
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
			ExpectBreaches: []breach.Breach{
				&breach.ValueBreach{
					BreachType: "value",
					CheckType:  "yaml",
					Severity:   "normal",
					ValueLabel: "- no file",
					Value:      "no matching yaml files found",
				},
			},
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
			ExpectPasses: []string{"no matching config files found"},
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
