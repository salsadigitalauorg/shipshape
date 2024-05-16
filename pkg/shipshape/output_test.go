package shipshape_test

import (
	"bufio"
	"bytes"
	"testing"
	"text/tabwriter"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	. "github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func TestTableDisplay(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	RunResultList = result.ResultList{}
	TableDisplay(w)
	assert.Equal(
		"No result available; ensure your shipshape.yml is configured correctly.\n",
		buf.String())

	buf = bytes.Buffer{}
	RunResultList = result.ResultList{Results: []result.Result{{Name: "a", Status: result.Pass}}}
	TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n", buf.String())

	buf = bytes.Buffer{}
	RunResultList = result.ResultList{
		Results: []result.Result{
			{Name: "a", Status: result.Pass},
			{Name: "b", Status: result.Pass},
			{Name: "c", Status: result.Pass},
		},
	}
	TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n"+
		"b      Pass              \n"+
		"c      Pass              \n",
		buf.String())

	buf = bytes.Buffer{}
	RunResultList = result.ResultList{
		Results: []result.Result{
			{
				Name:   "a",
				Status: result.Pass,
				Passes: []string{"Pass a", "Pass ab"},
			},
			{
				Name:   "b",
				Status: result.Pass,
				Passes: []string{"Pass b", "Pass bb", "Pass bc"},
			},
			{
				Name:   "c",
				Status: result.Fail,
				Breaches: []breach.Breach{
					&breach.ValueBreach{Value: "Fail c"},
					&breach.ValueBreach{Value: "Fail cb"},
				},
			},
			{
				Name:   "d",
				Status: result.Fail,
				Passes: []string{"Pass d", "Pass db"},
				Breaches: []breach.Breach{
					&breach.ValueBreach{Value: "Fail c"},
					&breach.ValueBreach{Value: "Fail cb"},
				},
			},
		},
	}
	TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES    FAILS\n"+
		"a      Pass     Pass a    \n"+
		"                Pass ab   \n"+
		"b      Pass     Pass b    \n"+
		"                Pass bb   \n"+
		"                Pass bc   \n"+
		"c      Fail               Fail c\n"+
		"                          Fail cb\n"+
		"d      Fail     Pass d    Fail c\n"+
		"                Pass db   Fail cb\n",
		buf.String())
}

func TestSimpleDisplay(t *testing.T) {
	assert := assert.New(t)

	t.Run("noResult", func(t *testing.T) {
		RunResultList = result.NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		SimpleDisplay(w)
		assert.Equal("No result available; ensure your shipshape.yml is configured correctly.\n", buf.String())
	})

	t.Run("topShape", func(t *testing.T) {
		RunResultList = result.NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name: "a", Status: result.Pass})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("breachesDetected", func(t *testing.T) {
		RunResultList = result.NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name:   "b",
			Status: result.Fail,
			Breaches: []breach.Breach{
				&breach.ValueBreach{Value: "Fail b"},
			},
		})
		SimpleDisplay(w)
		assert.Equal("# Breaches were detected\n\n  ### b\n     -- Fail b\n\n", buf.String())
	})

	t.Run("topShapeRemediating", func(t *testing.T) {
		RunResultList = result.ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.Results = append(RunResultList.Results, result.Result{
			Name: "a", Status: result.Pass})
		SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("allBreachesRemediated", func(t *testing.T) {
		RunResultList = result.ResultList{
			Results: []result.Result{{
				Name: "a",
				Breaches: []breach.Breach{
					&breach.ValueBreach{
						Remediation: breach.Remediation{
							Status:   breach.RemediationStatusSuccess,
							Messages: []string{"fixed 1"},
						},
					},
				}}},
			TotalBreaches:        1,
			RemediationPerformed: true,
			RemediationTotals:    map[string]uint32{"successful": 1},
		}

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		SimpleDisplay(w)
		assert.Equal("Breaches were detected but were all fixed successfully!\n\n"+
			"  ### a\n     -- fixed 1\n\n", buf.String())
	})

	t.Run("someBreachesRemediated", func(t *testing.T) {
		RunResultList = result.ResultList{
			Results: []result.Result{{
				Name: "a",
				Breaches: []breach.Breach{
					&breach.ValueBreach{
						Value: "Fail a",
						Remediation: breach.Remediation{
							Status:   breach.RemediationStatusSuccess,
							Messages: []string{"fixed 1"},
						},
					},
					&breach.ValueBreach{
						Value: "Fail b",
						Remediation: breach.Remediation{
							Status:   breach.RemediationStatusFailed,
							Messages: []string{"not fixed 1"},
						},
					},
				}}},
			TotalBreaches:        2,
			RemediationPerformed: true,
			RemediationTotals:    map[string]uint32{"successful": 1, "failed": 1},
		}

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		SimpleDisplay(w)
		assert.Equal("Breaches were detected but not all of them could be "+
			"fixed as they are either not supported yet or there were errors "+
			"when trying to remediate.\n\n"+
			"# Remediations\n\n  ### a\n     -- fixed 1\n\n"+
			"# Non-remediated breaches\n\n  ### a\n     -- Fail b\n\n", buf.String())
	})

	t.Run("noBreachRemediated", func(t *testing.T) {
		RunResultList = result.ResultList{
			Results: []result.Result{{
				Name: "a",
				Breaches: []breach.Breach{
					&breach.ValueBreach{
						Remediation: breach.Remediation{
							Status:   breach.RemediationStatusFailed,
							Messages: []string{"failed 1"},
						},
					},
				}}},
			TotalBreaches:        1,
			RemediationPerformed: true,
			RemediationTotals:    map[string]uint32{"failed": 1},
		}

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		SimpleDisplay(w)
		assert.Equal("Breaches were detected but none of them could be "+
			"fixed as there were errors when trying to remediate.\n\n"+
			"# Non-remediated breaches\n\n"+
			"  ### a\n     -- \n\n", buf.String())
	})
}

type testCheck struct{ config.CheckBase }

const testCheckType config.CheckType = "test-check"

func TestJUnit(t *testing.T) {
	assert := assert.New(t)

	RunResultList = result.NewResultList(false)
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0"></testsuites>
`, buf.String())

	RunConfig.Checks = config.CheckMap{testCheckType: []config.Check{&testCheck{
		CheckBase: config.CheckBase{Name: "a"},
	}}}
	RunResultList.Results = append(RunResultList.Results, result.Result{
		Name: "a", Status: result.Pass})
	buf = bytes.Buffer{}
	JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="test-check" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
    </testsuite>
</testsuites>
`, buf.String())

	RunConfig.Checks[testCheckType] = append(RunConfig.Checks[testCheckType], &testCheck{
		CheckBase: config.CheckBase{Name: "b"},
	})
	RunResultList.Results = append(RunResultList.Results, result.Result{
		Name:   "b",
		Status: result.Fail,
		Breaches: []breach.Breach{
			&breach.ValueBreach{Value: "Fail b"},
		},
	})
	buf = bytes.Buffer{}
	JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="test-check" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
        <testcase name="b" classname="b">
            <error message="Fail b"></error>
        </testcase>
    </testsuite>
</testsuites>
`, buf.String())
}
