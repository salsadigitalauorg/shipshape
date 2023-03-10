package config_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/config/testdata/testchecks"
	"github.com/salsadigitalauorg/shipshape/pkg/config/testdata/testchecks_invalid"
	"github.com/salsadigitalauorg/shipshape/pkg/crawler"
	"github.com/salsadigitalauorg/shipshape/pkg/file"
	shipshape_yaml "github.com/salsadigitalauorg/shipshape/pkg/yaml"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCheckMapUnmarshalYaml(t *testing.T) {
	assert := assert.New(t)

	t.Run("valid", func(t *testing.T) {
		var cm CheckMap
		testchecks.RegisterChecks()

		configBytes := []byte(`
test-check-1:
  - name: My test check 1
    foo: baz
test-check-2:
  - name: My first test check 2
    bar: zoom
  - name: My second test check 2
    bar: zap
`)
		err := yaml.Unmarshal(configBytes, &cm)
		assert.NoError(err)
		assert.Equal(CheckMap{
			testchecks.TestCheck1: {
				&testchecks.TestCheck1Check{
					CheckBase: CheckBase{Name: "My test check 1"},
					Foo:       "baz",
				},
			},
			testchecks.TestCheck2: {
				&testchecks.TestCheck2Check{
					CheckBase: CheckBase{Name: "My first test check 2"},
					Bar:       "zoom",
				},
				&testchecks.TestCheck2Check{
					CheckBase: CheckBase{Name: "My second test check 2"},
					Bar:       "zap",
				},
			},
		}, cm)
	})

	t.Run("nonSequenceNode", func(t *testing.T) {
		var cm CheckMap
		testchecks.RegisterChecks()

		configBytes := []byte(`
test-check-1:
  name: My test check 1
  foo: baz
`)
		err := yaml.Unmarshal(configBytes, &cm)
		assert.EqualError(err, "list required under check type 'test-check-1', got !!map instead")
	})

	t.Run("invalidYamlLookup", func(t *testing.T) {
		var cm CheckMap
		testchecks_invalid.RegisterChecks()
		configBytes := []byte(`
foo:
  - bar: baz
`)
		err := yaml.Unmarshal(configBytes, &cm)
		assert.EqualError(err, "invalid character ' ' at position 10, following \"test-check\"")
	})
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)

	cfg := Config{
		ProjectDir:   "foo",
		FailSeverity: NormalSeverity,
	}

	// Empty values should not change anything.
	err := cfg.Merge(Config{})
	assert.NoError(err)
	assert.Equal("foo", cfg.ProjectDir)
	assert.Equal(NormalSeverity, cfg.FailSeverity)

	// Ensure basic values are updated.
	err = cfg.Merge(Config{
		ProjectDir:   "bar",
		FailSeverity: HighSeverity,
	})
	assert.NoError(err)
	assert.Equal("bar", cfg.ProjectDir)
	assert.Equal(HighSeverity, cfg.FailSeverity)

	// Ensure checks are merged properly.
	err = cfg.Merge(Config{
		Checks: CheckMap{
			file.File: {&file.FileCheck{CheckBase: CheckBase{Name: "filecheck1"}}},
		},
	})
	assert.NoError(err)
	assert.Equal("bar", cfg.ProjectDir)
	assert.Equal(HighSeverity, cfg.FailSeverity)
	assert.EqualValues(
		CheckMap{
			file.File: {&file.FileCheck{CheckBase: CheckBase{Name: "filecheck1"}}},
		},
		cfg.Checks,
	)

	err = cfg.Merge(Config{
		Checks: CheckMap{
			shipshape_yaml.Yaml: {&shipshape_yaml.YamlCheck{
				YamlBase: shipshape_yaml.YamlBase{
					CheckBase: CheckBase{Name: "yamlcheck1"},
				},
			}},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			file.File: {&file.FileCheck{CheckBase: CheckBase{Name: "filecheck1"}}},
			shipshape_yaml.Yaml: {&shipshape_yaml.YamlCheck{
				YamlBase: shipshape_yaml.YamlBase{
					CheckBase: CheckBase{Name: "yamlcheck1"},
				},
			}},
		},
		cfg.Checks,
	)

	err = cfg.Merge(Config{
		Checks: CheckMap{
			crawler.Crawler: {&crawler.CrawlerCheck{CheckBase: CheckBase{Name: "crawlercheck1"}}},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			file.File: {&file.FileCheck{CheckBase: CheckBase{Name: "filecheck1"}}},
			shipshape_yaml.Yaml: {&shipshape_yaml.YamlCheck{
				YamlBase: shipshape_yaml.YamlBase{
					CheckBase: CheckBase{Name: "yamlcheck1"},
				},
			}},
			crawler.Crawler: {&crawler.CrawlerCheck{CheckBase: CheckBase{Name: "crawlercheck1"}}},
		},
		cfg.Checks,
	)

	cfg = Config{
		ProjectDir:   "foo",
		FailSeverity: NormalSeverity,
		Checks: CheckMap{
			file.File: {&file.FileCheck{
				CheckBase: CheckBase{Name: "filecheck1", Severity: NormalSeverity},
			}},
		},
	}
	err = cfg.Merge(Config{
		Checks: CheckMap{
			file.File: {&file.FileCheck{
				CheckBase: CheckBase{Name: "filecheck2", Severity: NormalSeverity},
				Path:      "path1"},
			},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			file.File: {
				&file.FileCheck{CheckBase: CheckBase{Name: "filecheck1", Severity: NormalSeverity}},
				&file.FileCheck{
					CheckBase: CheckBase{Name: "filecheck2", Severity: NormalSeverity},
					Path:      "path1"},
			},
		},
		cfg.Checks,
	)

	// Test changing values for same check name.
	err = cfg.Merge(Config{
		Checks: CheckMap{
			file.File: {
				&file.FileCheck{
					CheckBase: CheckBase{
						Name:     "filecheck2",
						Severity: HighSeverity},
					Path: "path2",
				},
			},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			file.File: {
				&file.FileCheck{
					CheckBase: CheckBase{
						Name:     "filecheck1",
						Severity: NormalSeverity,
					},
				},
				&file.FileCheck{
					CheckBase: CheckBase{
						Name:     "filecheck2",
						Severity: HighSeverity,
					},
					Path: "path2",
				},
			},
		},
		cfg.Checks,
	)

	// Test changing values for all checks of a type.
	err = cfg.Merge(Config{
		Checks: CheckMap{
			file.File: {
				&file.FileCheck{
					CheckBase: CheckBase{
						Severity: CriticalSeverity}}}}})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			file.File: {
				&file.FileCheck{
					CheckBase: CheckBase{
						Name:     "filecheck1",
						Severity: CriticalSeverity,
					},
				},
				&file.FileCheck{
					CheckBase: CheckBase{
						Name:     "filecheck2",
						Severity: CriticalSeverity,
					},
					Path: "path2",
				},
			},
		},
		cfg.Checks,
	)
}
