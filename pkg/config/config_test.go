package config_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/config/testdata/filterchecks"
	"github.com/salsadigitalauorg/shipshape/pkg/config/testdata/testchecks"
	"github.com/salsadigitalauorg/shipshape/pkg/config/testdata/testchecks_invalid"

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

	testchecks.RegisterChecks()

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
			testchecks.TestCheck1: {&testchecks.TestCheck1Check{CheckBase: CheckBase{Name: "filecheck1"}}},
		},
	})
	assert.NoError(err)
	assert.Equal("bar", cfg.ProjectDir)
	assert.Equal(HighSeverity, cfg.FailSeverity)
	assert.EqualValues(
		CheckMap{
			testchecks.TestCheck1: {&testchecks.TestCheck1Check{CheckBase: CheckBase{Name: "filecheck1"}}},
		},
		cfg.Checks,
	)

	err = cfg.Merge(Config{
		Checks: CheckMap{
			testchecks.TestCheck2: {&testchecks.TestCheck2Check{CheckBase: CheckBase{Name: "yamlcheck1"}}},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			testchecks.TestCheck1: {&testchecks.TestCheck1Check{CheckBase: CheckBase{Name: "filecheck1"}}},
			testchecks.TestCheck2: {&testchecks.TestCheck2Check{CheckBase: CheckBase{Name: "yamlcheck1"}}},
		},
		cfg.Checks,
	)

	err = cfg.Merge(Config{
		Checks: CheckMap{
			testchecks.TestCheck3: {&testchecks.TestCheck3Check{CheckBase: CheckBase{Name: "crawlercheck1"}}},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			testchecks.TestCheck1: {&testchecks.TestCheck1Check{CheckBase: CheckBase{Name: "filecheck1"}}},
			testchecks.TestCheck2: {&testchecks.TestCheck2Check{CheckBase: CheckBase{Name: "yamlcheck1"}}},
			testchecks.TestCheck3: {&testchecks.TestCheck3Check{CheckBase: CheckBase{Name: "crawlercheck1"}}},
		},
		cfg.Checks,
	)

	cfg = Config{
		ProjectDir:   "foo",
		FailSeverity: NormalSeverity,
		Checks: CheckMap{
			testchecks.TestCheck1: {&testchecks.TestCheck1Check{
				CheckBase: CheckBase{Name: "filecheck1", Severity: NormalSeverity},
			}},
		},
	}
	err = cfg.Merge(Config{
		Checks: CheckMap{
			testchecks.TestCheck1: {&testchecks.TestCheck1Check{
				CheckBase: CheckBase{Name: "filecheck2", Severity: NormalSeverity},
				Foo:       "path1"},
			},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			testchecks.TestCheck1: {
				&testchecks.TestCheck1Check{
					CheckBase: CheckBase{Name: "filecheck1", Severity: NormalSeverity}},
				&testchecks.TestCheck1Check{
					CheckBase: CheckBase{Name: "filecheck2", Severity: NormalSeverity},
					Foo:       "path1"},
			},
		},
		cfg.Checks,
	)

	// Test changing values for same check name.
	err = cfg.Merge(Config{
		Checks: CheckMap{
			testchecks.TestCheck1: {
				&testchecks.TestCheck1Check{
					CheckBase: CheckBase{
						Name:     "filecheck2",
						Severity: HighSeverity},
					Foo: "path2",
				},
			},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			testchecks.TestCheck1: {
				&testchecks.TestCheck1Check{
					CheckBase: CheckBase{
						Name:     "filecheck1",
						Severity: NormalSeverity,
					},
				},
				&testchecks.TestCheck1Check{
					CheckBase: CheckBase{
						Name:     "filecheck2",
						Severity: HighSeverity,
					},
					Foo: "path2",
				},
			},
		},
		cfg.Checks,
	)

	// Test changing values for all checks of a type.
	err = cfg.Merge(Config{
		Checks: CheckMap{
			testchecks.TestCheck1: {
				&testchecks.TestCheck1Check{
					CheckBase: CheckBase{
						Severity: CriticalSeverity}}}}})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			testchecks.TestCheck1: {
				&testchecks.TestCheck1Check{
					CheckBase: CheckBase{
						Name:     "filecheck1",
						Severity: CriticalSeverity,
					},
				},
				&testchecks.TestCheck1Check{
					CheckBase: CheckBase{
						Name:     "filecheck2",
						Severity: CriticalSeverity,
					},
					Foo: "path2",
				},
			},
		},
		cfg.Checks,
	)
}

func TestFilterChecksToRun(t *testing.T) {
	assert := assert.New(t)

	t.Run("filterByCheckTypes", func(t *testing.T) {
		cfg := Config{
			Checks: CheckMap{
				filterchecks.FilterCheck1: {
					&filterchecks.FilterCheck1Check{
						CheckBase: CheckBase{Name: "filter check 1"},
					},
				},
				filterchecks.FilterCheck2: {
					&filterchecks.FilterCheck2Check{
						CheckBase: CheckBase{Name: "filter check 2"},
					},
				},
			},
		}
		for ct, checks := range cfg.Checks {
			for _, c := range checks {
				c.Init(ct)
			}
		}
		cfg.FilterChecksToRun([]string{"filter-check-1"}, false)

		expectedCheck := &filterchecks.FilterCheck1Check{
			CheckBase: CheckBase{Name: "filter check 1"},
		}
		expectedCheck.Init(filterchecks.FilterCheck1)
		assert.EqualValues(Config{
			Checks: CheckMap{filterchecks.FilterCheck1: {expectedCheck}},
		}, cfg)
	})

	t.Run("filterOutDb", func(t *testing.T) {
		cfg := Config{
			Checks: CheckMap{
				filterchecks.FilterCheck1: {
					&filterchecks.FilterCheck1Check{
						CheckBase: CheckBase{Name: "filter check 1"},
					},
				},
				filterchecks.FilterCheck2: {
					&filterchecks.FilterCheck2Check{
						CheckBase: CheckBase{Name: "filter check 2"},
					},
				},
			},
		}
		for ct, checks := range cfg.Checks {
			for _, c := range checks {
				c.Init(ct)
			}
		}
		cfg.FilterChecksToRun([]string(nil), true)

		expectedCheck := &filterchecks.FilterCheck2Check{
			CheckBase: CheckBase{Name: "filter check 2"},
		}
		expectedCheck.Init(filterchecks.FilterCheck2)
		assert.EqualValues(Config{
			Checks: CheckMap{filterchecks.FilterCheck2: {expectedCheck}},
		}, cfg)
	})
}
