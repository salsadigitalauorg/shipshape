package config_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/crawler"
	"github.com/salsadigitalauorg/shipshape/pkg/file"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"

	"github.com/stretchr/testify/assert"
)

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
			shipshape.Yaml: {&shipshape.YamlCheck{
				YamlBase: shipshape.YamlBase{
					CheckBase: CheckBase{Name: "yamlcheck1"},
				},
			}},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		CheckMap{
			file.File: {&file.FileCheck{CheckBase: CheckBase{Name: "filecheck1"}}},
			shipshape.Yaml: {&shipshape.YamlCheck{
				YamlBase: shipshape.YamlBase{
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
			shipshape.Yaml: {&shipshape.YamlCheck{
				YamlBase: shipshape.YamlBase{
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
