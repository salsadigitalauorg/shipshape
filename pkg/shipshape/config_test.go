package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMerge(t *testing.T) {
	assert := assert.New(t)

	cfg := shipshape.Config{
		ProjectDir:   "foo",
		FailSeverity: shipshape.NormalSeverity,
	}

	// Empty values should not change anything.
	err := cfg.Merge(shipshape.Config{})
	assert.NoError(err)
	assert.Equal("foo", cfg.ProjectDir)
	assert.Equal(shipshape.NormalSeverity, cfg.FailSeverity)

	// Ensure basic values are updated.
	err = cfg.Merge(shipshape.Config{
		ProjectDir:   "bar",
		FailSeverity: shipshape.HighSeverity,
	})
	assert.NoError(err)
	assert.Equal("bar", cfg.ProjectDir)
	assert.Equal(shipshape.HighSeverity, cfg.FailSeverity)

	// Ensure checks are merged properly.
	err = cfg.Merge(shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{CheckBase: shipshape.CheckBase{Name: "filecheck1"}}},
		},
	})
	assert.NoError(err)
	assert.Equal("bar", cfg.ProjectDir)
	assert.Equal(shipshape.HighSeverity, cfg.FailSeverity)
	assert.EqualValues(
		shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{CheckBase: shipshape.CheckBase{Name: "filecheck1"}}},
		},
		cfg.Checks,
	)

	err = cfg.Merge(shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.Yaml: {&shipshape.YamlCheck{
				YamlBase: shipshape.YamlBase{
					CheckBase: shipshape.CheckBase{Name: "yamlcheck1"},
				},
			}},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{CheckBase: shipshape.CheckBase{Name: "filecheck1"}}},
			shipshape.Yaml: {&shipshape.YamlCheck{
				YamlBase: shipshape.YamlBase{
					CheckBase: shipshape.CheckBase{Name: "yamlcheck1"},
				},
			}},
		},
		cfg.Checks,
	)

	err = cfg.Merge(shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.Crawler: {&shipshape.CrawlerCheck{CheckBase: shipshape.CheckBase{Name: "crawlercheck1"}}},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{CheckBase: shipshape.CheckBase{Name: "filecheck1"}}},
			shipshape.Yaml: {&shipshape.YamlCheck{
				YamlBase: shipshape.YamlBase{
					CheckBase: shipshape.CheckBase{Name: "yamlcheck1"},
				},
			}},
			shipshape.Crawler: {&shipshape.CrawlerCheck{CheckBase: shipshape.CheckBase{Name: "crawlercheck1"}}},
		},
		cfg.Checks,
	)

	cfg = shipshape.Config{
		ProjectDir:   "foo",
		FailSeverity: shipshape.NormalSeverity,
		Checks: shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{
				CheckBase: shipshape.CheckBase{Name: "filecheck1", Severity: shipshape.NormalSeverity},
			}},
		},
	}
	err = cfg.Merge(shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{
				CheckBase: shipshape.CheckBase{Name: "filecheck2", Severity: shipshape.NormalSeverity},
				Path:      "path1"},
			},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		shipshape.CheckMap{
			shipshape.File: {
				&shipshape.FileCheck{CheckBase: shipshape.CheckBase{Name: "filecheck1", Severity: shipshape.NormalSeverity}},
				&shipshape.FileCheck{
					CheckBase: shipshape.CheckBase{Name: "filecheck2", Severity: shipshape.NormalSeverity},
					Path:      "path1"},
			},
		},
		cfg.Checks,
	)

	// Test changing values for same check name.
	err = cfg.Merge(shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.File: {
				&shipshape.FileCheck{
					CheckBase: shipshape.CheckBase{
						Name:     "filecheck2",
						Severity: shipshape.HighSeverity},
					Path: "path2",
				},
			},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		shipshape.CheckMap{
			shipshape.File: {
				&shipshape.FileCheck{
					CheckBase: shipshape.CheckBase{
						Name:     "filecheck1",
						Severity: shipshape.NormalSeverity,
					},
				},
				&shipshape.FileCheck{
					CheckBase: shipshape.CheckBase{
						Name:     "filecheck2",
						Severity: shipshape.HighSeverity,
					},
					Path: "path2",
				},
			},
		},
		cfg.Checks,
	)

	// Test changing values for all checks of a type.
	err = cfg.Merge(shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.File: {
				&shipshape.FileCheck{
					CheckBase: shipshape.CheckBase{
						Severity: shipshape.CriticalSeverity}}}}})
	assert.NoError(err)
	assert.EqualValues(
		shipshape.CheckMap{
			shipshape.File: {
				&shipshape.FileCheck{
					CheckBase: shipshape.CheckBase{
						Name:     "filecheck1",
						Severity: shipshape.CriticalSeverity,
					},
				},
				&shipshape.FileCheck{
					CheckBase: shipshape.CheckBase{
						Name:     "filecheck2",
						Severity: shipshape.CriticalSeverity,
					},
					Path: "path2",
				},
			},
		},
		cfg.Checks,
	)
}

func TestRunChecks(t *testing.T) {
	assert := assert.New(t)

	testFileCheck := &shipshape.FileCheck{}
	testYamlCheck := &shipshape.FileCheck{}
	yaml.Unmarshal([]byte("name: filecheck1"), testFileCheck)
	testFileCheck.Init(shipshape.File)
	yaml.Unmarshal([]byte("name: yamlcheck1"), testYamlCheck)
	testYamlCheck.Init(shipshape.Yaml)
	cfg := shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.File: {testFileCheck},
			shipshape.Yaml: {testYamlCheck},
		},
	}

	rl := cfg.RunChecks()
	assert.Equal(uint32(2), rl.TotalChecks)
	assert.Equal(uint32(2), rl.TotalBreaches)
	assert.EqualValues(map[shipshape.CheckType]int{
		shipshape.File: 1,
		shipshape.Yaml: 1,
	}, rl.BreachCountByType)
	assert.ElementsMatch([]shipshape.Result{
		{Name: "filecheck1", Severity: "normal", CheckType: "file", Status: "Fail", Passes: []string(nil), Failures: []string{"directory not provided"}, Warnings: []string(nil)},
		{Name: "yamlcheck1", Severity: "normal", CheckType: "yaml", Status: "Fail", Passes: []string(nil), Failures: []string{"directory not provided"}, Warnings: []string(nil)}},
		rl.Results)
}
