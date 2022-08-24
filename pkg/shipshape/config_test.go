package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
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
			shipshape.File: {&shipshape.FileCheck{}},
			shipshape.Yaml: {&shipshape.YamlCheck{}},
		},
	})
	assert.NoError(err)
	assert.Equal("bar", cfg.ProjectDir)
	assert.Equal(shipshape.HighSeverity, cfg.FailSeverity)
	assert.EqualValues(
		shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{}},
			shipshape.Yaml: {&shipshape.YamlCheck{}},
		},
		cfg.Checks,
	)

	// Adding the same values should be skipped.
	err = cfg.Merge(shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{}},
			shipshape.Yaml: {&shipshape.YamlCheck{}},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{}},
			shipshape.Yaml: {&shipshape.YamlCheck{}},
		},
		cfg.Checks,
	)

	err = cfg.Merge(shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.Crawler: {&shipshape.CrawlerCheck{}},
		},
	})
	assert.NoError(err)
	assert.EqualValues(
		shipshape.CheckMap{
			shipshape.File:    {&shipshape.FileCheck{}},
			shipshape.Yaml:    {&shipshape.YamlCheck{}},
			shipshape.Crawler: {&shipshape.CrawlerCheck{}},
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
}

func TestRunChecks(t *testing.T) {
	cfg := shipshape.Config{
		Checks: shipshape.CheckMap{
			shipshape.File: {&shipshape.FileCheck{}},
			shipshape.Yaml: {&shipshape.YamlCheck{}},
		},
	}
	cfg.RunChecks()
}
