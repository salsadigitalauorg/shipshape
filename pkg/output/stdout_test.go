package output_test

import (
	"bufio"
	"bytes"
	"testing"
	"text/tabwriter"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestTableDisplay(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	s := &Stdout{Format: "table", ResultList: &result.ResultList{}}
	s.Table(w)
	assert.Equal(
		"No result available; ensure your shipshape.yml is configured correctly.\n",
		buf.String())

	buf = bytes.Buffer{}
	s = &Stdout{Format: "table", ResultList: &result.ResultList{
		Results: []result.Result{{Name: "a", Status: result.Pass}}}}
	s.Table(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n", buf.String())

	buf = bytes.Buffer{}
	s = &Stdout{Format: "table", ResultList: &result.ResultList{
		Results: []result.Result{
			{Name: "a", Status: result.Pass},
			{Name: "b", Status: result.Pass},
			{Name: "c", Status: result.Pass},
		},
	}}
	s.Table(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n"+
		"b      Pass              \n"+
		"c      Pass              \n",
		buf.String())

	buf = bytes.Buffer{}
	s = &Stdout{Format: "table", ResultList: &result.ResultList{
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
	}}
	s.Table(w)
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

func TestPrettyDisplay(t *testing.T) {
	assert := assert.New(t)

	t.Run("noResult", func(t *testing.T) {
		rl := result.NewResultList(false)

		s := &Stdout{Format: "pretty", ResultList: &rl}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		s.Pretty(w)
		assert.Equal("No result available; ensure your shipshape.yml is configured correctly.\n", buf.String())
	})

	t.Run("topShape", func(t *testing.T) {
		rl := result.NewResultList(false)
		rl.Results = append(rl.Results, result.Result{
			Name: "a", Status: result.Pass})

		s := &Stdout{Format: "pretty", ResultList: &rl}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		buf = bytes.Buffer{}
		s.Pretty(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("breachesDetected", func(t *testing.T) {
		rl := result.NewResultList(false)
		rl.Results = append(rl.Results, result.Result{
			Name:   "b",
			Status: result.Fail,
			Breaches: []breach.Breach{
				&breach.ValueBreach{Value: "Fail b"},
			},
		})

		s := &Stdout{Format: "pretty", ResultList: &rl}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		s.Pretty(w)
		assert.Equal("# Breaches were detected\n\n  ### b\n     -- Fail b\n\n", buf.String())
	})

	t.Run("topShapeRemediating", func(t *testing.T) {
		rl := result.NewResultList(true)
		rl.Results = append(rl.Results, result.Result{
			Name: "a", Status: result.Pass})

		s := &Stdout{Format: "pretty", ResultList: &rl}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		s.Pretty(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("allBreachesRemediated", func(t *testing.T) {
		rl := result.ResultList{
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

		s := &Stdout{Format: "pretty", ResultList: &rl}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		s.Pretty(w)
		assert.Equal("Breaches were detected but were all fixed successfully!\n\n"+
			"  ### a\n     -- fixed 1\n\n", buf.String())
	})

	t.Run("someBreachesRemediated", func(t *testing.T) {
		rl := result.ResultList{
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

		s := &Stdout{Format: "pretty", ResultList: &rl}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		s.Pretty(w)
		assert.Equal("Breaches were detected but not all of them could be "+
			"fixed as they are either not supported yet or there were errors "+
			"when trying to remediate.\n\n"+
			"# Remediations\n\n  ### a\n     -- fixed 1\n\n"+
			"# Non-remediated breaches\n\n  ### a\n     -- Fail b\n\n", buf.String())
	})

	t.Run("noBreachRemediated", func(t *testing.T) {
		rl := result.ResultList{
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

		s := &Stdout{Format: "pretty", ResultList: &rl}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		s.Pretty(w)
		assert.Equal("Breaches were detected but none of them could be "+
			"fixed as there were errors when trying to remediate.\n\n"+
			"# Non-remediated breaches\n\n"+
			"  ### a\n     -- \n\n", buf.String())
	})
}

func TestJUnit(t *testing.T) {
	assert := assert.New(t)

	rl := result.NewResultList(false)
	s := &Stdout{Format: "junit", ResultList: &rl}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	s.JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0"></testsuites>
`, buf.String())

	rl.Policies = map[string][]string{"test-check": {"a"}}
	rl.Results = append(rl.Results, result.Result{
		Name: "a", Status: result.Pass})
	buf = bytes.Buffer{}
	s.JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="test-check" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
    </testsuite>
</testsuites>
`, buf.String())

	rl.Policies["test-check"] = append(rl.Policies["test-check"], "b")
	rl.Results = append(rl.Results, result.Result{
		Name:   "b",
		Status: result.Fail,
		Breaches: []breach.Breach{
			&breach.ValueBreach{Value: "Fail b"},
		},
	})
	buf = bytes.Buffer{}
	s.JUnit(w)
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
