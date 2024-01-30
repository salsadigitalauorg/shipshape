package yaml_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
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
	tt := []struct {
		name             string
		check            YamlCheck
		expectedBreaches []result.Breach
		expectedPasses   []string
		expectResultFail bool
		expectResultPass bool
		expectDataMap    map[string][]byte
	}{
		{
			name: "noFile",
			check: YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
			},
			expectedBreaches: []result.Breach{
				result.ValueBreach{
					BreachType: "value",
					ValueLabel: "- no file",
					Value:      "no file provided",
				},
			},
			expectResultFail: true,
		},

		{
			name: "nonExistentFile",
			check: YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				File: "non-existent.yml",
			},
			expectedBreaches: []result.Breach{
				result.ValueBreach{
					BreachType: "value",
					ValueLabel: "error reading file: testdata/non-existent.yml",
					Value:      "open testdata/non-existent.yml: no such file or directory",
				},
			},
			expectResultFail: true,
		},

		{
			name: "nonExistentFileIgnoreMissing",
			check: YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				File:          "non-existent.yml",
				IgnoreMissing: &cTrue,
			},
			expectedPasses:   []string{"File testdata/non-existent.yml does not exist"},
			expectResultPass: true,
		},

		{
			name: "singleFile",
			check: YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				File: "update.settings.yml",
			},
			expectDataMap: map[string][]byte{
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
			name: "badFilePattern",
			check: YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern: "*.bar.yml",
				Path:    "",
			},
			expectedBreaches: []result.Breach{
				result.ValueBreach{
					BreachType: "value",
					ValueLabel: "error finding files in path: testdata",
					Value:      "error parsing regexp: missing argument to repetition operator: `*`",
				},
			},
			expectResultFail: true,
		},

		{
			name: "filePatternNoMatchingFile",
			check: YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern: "bla.*.yml",
			},
			expectedBreaches: []result.Breach{
				result.ValueBreach{
					BreachType: "value",
					ValueLabel: "- no file",
					Value:      "no matching yaml files found",
				},
			},
			expectResultFail: true,
		},

		{
			name: "filePatternNoMatchingFileIgnoreMissing",
			check: YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern:       "bla.*.yml",
				IgnoreMissing: &cTrue,
			},
			expectedPasses:   []string{"no matching config files found"},
			expectResultPass: true,
		},

		{
			name: "correctSingleFilePatternAndValue",
			check: YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern: "foo.bar.yml",
				Path:    "dir/subdir",
			},
			expectDataMap: map[string][]byte{
				"testdata/dir/subdir/foo.bar.yml": []byte(
					`check:
  interval_days: 7
`),
			},
		},

		{
			name: "recursiveFileLookup",
			check: YamlCheck{
				YamlBase: YamlBase{
					Values: []KeyValue{
						{Key: "check.interval_days", Value: "7"},
					},
				},
				Pattern: ".*.bar.yml",
			},
			expectDataMap: map[string][]byte{
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
		t.Run(tc.name, func(innerT *testing.T) {
			assert := assert.New(innerT)
			tc.check.FetchData()

			if tc.expectResultFail {
				assert.Equal(result.Fail, tc.check.Result.Status)
			} else if tc.expectResultPass {
				assert.Equal(result.Pass, tc.check.Result.Status)
			} else {
				assert.NotEqual(result.Fail, tc.check.Result.Status)
				assert.NotEqual(result.Pass, tc.check.Result.Status)
			}

			if len(tc.expectedPasses) > 0 {
				assert.ElementsMatch(tc.expectedPasses, tc.check.Result.Passes)
			} else {
				assert.Empty(tc.check.Result.Passes)
			}

			if len(tc.expectedBreaches) > 0 {
				assert.ElementsMatch(tc.expectedBreaches, tc.check.Result.Breaches)
			} else {
				assert.Empty(tc.check.Result.Breaches)
			}

			if tc.expectDataMap != nil {
				assert.EqualValues(tc.expectDataMap, tc.check.DataMap)
			}
		})
	}
}
