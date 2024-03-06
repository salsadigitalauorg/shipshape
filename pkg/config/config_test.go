package config_test

import (
	"io"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	. "github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/config/testdata/filterchecks"
	"github.com/salsadigitalauorg/shipshape/pkg/config/testdata/testchecks"
	"github.com/salsadigitalauorg/shipshape/pkg/config/testdata/testchecks_invalid"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestReadAndParseConfig(t *testing.T) {
	assert := assert.New(t)

	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	t.Run("nonExistentFile", func(t *testing.T) {
		config.Files = []string{"testdata/nonexistent.yml"}
		_, _, _, err := ReadAndParseConfig()
		assert.Error(err)
		assert.Equal("open testdata/nonexistent.yml: no such file or directory", err.Error())
	})

	t.Run("existingFile", func(t *testing.T) {
		config.Files = []string{"testdata/shipshape.yml"}
		_, _, _, err := ReadAndParseConfig()
		assert.NoError(err)
	})
}

func TestParseConfigData(t *testing.T) {
	assert := assert.New(t)

	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)

	t.Run("invalidData", func(t *testing.T) {
		testchecks.RegisterChecks()
		logrus.SetOutput(io.Discard)
		invalidData := `
checks:
  test-check-1: foo
`
		_, _, _, err := ParseConfigData([][]byte{[]byte(invalidData)})
		assert.EqualError(err, "list required under check type 'test-check-1', got !!str instead")

	})

	t.Run("validData", func(t *testing.T) {
		testchecks.RegisterChecks()
		data := `
checks:
  test-check-1:
    - name: My test check 1
      foo: baz
  test-check-2:
    - name: My first test check 2
      bar: zoom
    - name: My second test check 2
      bar: zap
`
		_, cfg, _, err := ParseConfigData([][]byte{[]byte(data)})
		assert.NoError(err)

		if !assert.Len(cfg.Checks[testchecks.TestCheck1], 1) {
			t.FailNow()
		}
		if !assert.Len(cfg.Checks[testchecks.TestCheck2], 2) {
			t.FailNow()
		}

		tc1, ok := cfg.Checks[testchecks.TestCheck1][0].(*testchecks.TestCheck1Check)
		assert.True(ok)
		assert.Equal("My test check 1", tc1.Name)
		assert.Equal("baz", tc1.Foo)

		tc2, ok := cfg.Checks[testchecks.TestCheck2][0].(*testchecks.TestCheck2Check)
		assert.True(ok)
		assert.Equal("My first test check 2", tc2.Name)
		assert.Equal("zoom", tc2.Bar)

		tc22, ok := cfg.Checks[testchecks.TestCheck2][1].(*testchecks.TestCheck2Check)
		assert.True(ok)
		assert.Equal("My second test check 2", tc22.Name)
		assert.Equal("zap", tc22.Bar)
	})
}

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

	cfg := Config{}

	// Empty values should not change anything.
	err := cfg.Merge(Config{})
	assert.NoError(err)

	// Ensure checks are merged properly.
	err = cfg.Merge(Config{
		Checks: CheckMap{
			testchecks.TestCheck1: {&testchecks.TestCheck1Check{CheckBase: CheckBase{Name: "filecheck1"}}},
		},
	})
	assert.NoError(err)
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
		config.CheckTypesToRun = []string{"filter-check-1"}
		config.ExcludeDb = false
		cfg.FilterChecksToRun()

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
		config.CheckTypesToRun = []string(nil)
		config.ExcludeDb = true
		cfg.FilterChecksToRun()

		expectedCheck := &filterchecks.FilterCheck2Check{
			CheckBase: CheckBase{Name: "filter check 2"},
		}
		expectedCheck.Init(filterchecks.FilterCheck2)
		assert.EqualValues(Config{
			Checks: CheckMap{filterchecks.FilterCheck2: {expectedCheck}},
		}, cfg)
	})
}
