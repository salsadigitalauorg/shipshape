package yaml_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestYamlLintMerge(t *testing.T) {
	assert := assert.New(t)

	c := YamlLintCheck{
		YamlCheck: YamlCheck{
			Path: "/path/to/run/check",
		},
	}
	c.Merge(&YamlLintCheck{
		YamlCheck: YamlCheck{
			Path: "/new/path/to/run/check",
		},
	})
	assert.EqualValues(YamlLintCheck{
		YamlCheck: YamlCheck{
			Path: "/new/path/to/run/check",
		},
	}, c)
}

func TestYamlLintCheck(t *testing.T) {
	assert := assert.New(t)

	mockCheck := func(file string, files []string, ignoreMissing bool) YamlLintCheck {
		return YamlLintCheck{
			YamlCheck: YamlCheck{
				YamlBase: YamlBase{
					CheckBase: config.CheckBase{
						Name:    "Test yaml lint",
						DataMap: map[string][]byte{},
					},
				},
				File:          file,
				Files:         files,
				IgnoreMissing: &ignoreMissing,
			},
		}
	}

	c := mockCheck("", []string{}, false)
	c.Init(YamlLint)
	c.FetchData()
	assert.Equal(result.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.ElementsMatch(
		[]result.Breach{
			&result.ValueBreach{
				BreachType: "value",
				CheckType:  "yamllint",
				CheckName:  "Test yaml lint",
				Severity:   "normal",
				ValueLabel: "Test yaml lint- no file",
				Value:      "no file provided",
			},
		},
		c.Result.Breaches,
	)

	c = mockCheck("non-existent-file.yml", []string{}, true)
	c.Init(YamlLint)
	c.FetchData()
	assert.NotEqual(result.Fail, c.Result.Status)
	assert.Empty(c.Result.Breaches)
	assert.ElementsMatch(
		[]string{"File testdata/non-existent-file.yml does not exist"},
		c.Result.Passes,
	)

	c = mockCheck("", []string{"non-existent-file.yml", "yaml-invalid.yml"}, true)
	c.Init(YamlLint)
	c.FetchData()
	assert.NotEqual(result.Fail, c.Result.Status)
	assert.Empty(c.Result.Breaches)
	assert.ElementsMatch([]string{
		"File testdata/non-existent-file.yml does not exist",
		"File testdata/yaml-invalid.yml does not exist",
	}, c.Result.Passes)

	c = mockCheck("non-existent-file.yml", []string{}, false)
	c.Init(YamlLint)
	c.FetchData()
	assert.Equal(result.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.ElementsMatch(
		[]result.Breach{
			&result.ValueBreach{
				BreachType: "value",
				CheckType:  "yamllint",
				CheckName:  "Test yaml lint",
				Severity:   "normal",
				ValueLabel: "error reading file: testdata/non-existent-file.yml",
				Value:      "open testdata/non-existent-file.yml: no such file or directory",
			},
		},
		c.Result.Breaches,
	)

	c = mockCheck("", []string{"non-existent-file.yml", "yamllint-invalid.yml"}, false)
	c.Init(YamlLint)
	c.FetchData()
	assert.Equal(result.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.ElementsMatch(
		[]result.Breach{
			&result.ValueBreach{
				BreachType: "value",
				CheckType:  "yamllint",
				CheckName:  "Test yaml lint",
				Severity:   "normal",
				ValueLabel: "error reading file: testdata/non-existent-file.yml",
				Value:      "open testdata/non-existent-file.yml: no such file or directory",
			},
		},
		c.Result.Breaches,
	)

	c = mockCheck("", []string{}, false)
	c.Init(YamlLint)
	c.DataMap["yaml-invalid.yml"] = []byte(`
this: is invalid
this: yaml
`)
	c.UnmarshalDataMap()
	assert.Equal(result.Fail, c.Result.Status)
	assert.Empty(c.Result.Passes)
	assert.ElementsMatch(
		[]result.Breach{
			&result.ValueBreach{
				BreachType: "value",
				CheckType:  "yamllint",
				CheckName:  "Test yaml lint",
				Severity:   "normal",
				ValueLabel: "cannot decode yaml: yaml-invalid.yml",
				Value:      "line 3: mapping key \"this\" already defined at line 2",
			},
		},
		c.Result.Breaches,
	)

	c = mockCheck("", []string{}, false)
	c.Init(YamlLint)
	c.DataMap["yaml-valid.yml"] = []byte(`
this: is
valid: yaml
`)
	c.UnmarshalDataMap()
	assert.Equal(result.Pass, c.Result.Status)
	assert.Empty(c.Result.Breaches)
	assert.ElementsMatch(
		[]string{"yaml-valid.yml has valid yaml."},
		c.Result.Passes,
	)

	t.Run("rootInvalid", func(t *testing.T) {
		c := YamlLintCheck{}
		c.DataMap = map[string][]byte{
			"yaml-invalid-root.yml": []byte(`
foo: bar
- item 1
`)}
		c.UnmarshalDataMap()
		assert.Equal(result.Fail, c.Result.Status)
		assert.Empty(c.Result.Passes)
		assert.ElementsMatch(
			[]result.Breach{
				&result.ValueBreach{
					BreachType: "value",
					ValueLabel: "yaml error: yaml-invalid-root.yml",
					Value:      "yaml: line 1: did not find expected key",
				},
			},
			c.Result.Breaches,
		)
	})

	t.Run("rootList", func(t *testing.T) {
		c := YamlLintCheck{}
		c.DataMap = map[string][]byte{
			"yaml-valid-list.yml": []byte(`
- item 1
- item 2:
    foo: bar
`)}
		c.UnmarshalDataMap()
		assert.Equal(result.Pass, c.Result.Status)
		assert.Empty(c.Result.Breaches)
		assert.ElementsMatch(
			[]string{"yaml-valid-list.yml has valid yaml."},
			c.Result.Passes,
		)
	})
}
