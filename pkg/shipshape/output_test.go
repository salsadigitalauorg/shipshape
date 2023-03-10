package shipshape_test

import (
	"bufio"
	"bytes"
	"testing"
	"text/tabwriter"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/file"
	. "github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestTableDisplay(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	RunResultList = config.ResultList{}
	TableDisplay(w)
	assert.Equal(
		"No result available; ensure your shipshape.yml is configured correctly.\n",
		buf.String())

	buf = bytes.Buffer{}
	RunResultList = config.ResultList{Results: []config.Result{{Name: "a", Status: config.Pass}}}
	TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n", buf.String())

	buf = bytes.Buffer{}
	RunResultList = config.ResultList{
		Results: []config.Result{
			{Name: "a", Status: config.Pass},
			{Name: "b", Status: config.Pass},
			{Name: "c", Status: config.Pass},
		},
	}
	TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n"+
		"b      Pass              \n"+
		"c      Pass              \n",
		buf.String())

	buf = bytes.Buffer{}
	RunResultList = config.ResultList{
		Results: []config.Result{
			{
				Name:   "a",
				Status: config.Pass,
				Passes: []string{"Pass a", "Pass ab"},
			},
			{
				Name:   "b",
				Status: config.Pass,
				Passes: []string{"Pass b", "Pass bb", "Pass bc"},
			},
			{
				Name:     "c",
				Status:   config.Fail,
				Failures: []string{"Fail c", "Fail cb"},
			},
			{
				Name:     "d",
				Status:   config.Fail,
				Passes:   []string{"Pass d", "Pass db"},
				Failures: []string{"Fail c", "Fail cb"},
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
		RunResultList = config.NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		SimpleDisplay(w)
		assert.Equal("No result available; ensure your shipshape.yml is configured correctly.\n", buf.String())
	})

	t.Run("topShape", func(t *testing.T) {
		RunResultList = config.NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.Results = append(RunResultList.Results, config.Result{Name: "a", Status: config.Pass})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("breachesDetected", func(t *testing.T) {
		RunResultList = config.NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.Results = append(RunResultList.Results, config.Result{
			Name:     "b",
			Status:   config.Fail,
			Failures: []string{"Fail b"}})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("# Breaches were detected\n\n  ### b\n     -- Fail b\n\n", buf.String())
	})

	t.Run("topShapeRemediating", func(t *testing.T) {
		RunResultList = config.ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.Results = append(RunResultList.Results, config.Result{Name: "a", Status: config.Pass})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("allBreachesRemediated", func(t *testing.T) {
		RunResultList = config.ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.TotalRemediations = 1
		RunResultList.Results = append(RunResultList.Results, config.Result{Name: "a", Status: config.Pass, Remediations: []string{"fixed 1"}})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Breaches were detected but were all fixed successfully!\n\n"+
			"  ### a\n     -- fixed 1\n\n", buf.String())
	})

	t.Run("someBreachesRemediated", func(t *testing.T) {
		RunResultList = config.ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.TotalRemediations = 1
		RunResultList.TotalBreaches = 1
		RunResultList.Results = append(RunResultList.Results, config.Result{Name: "a", Status: config.Fail, Remediations: []string{"fixed 1"}})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Breaches were detected but not all of them could be "+
			"fixed as they are either not supported yet or there were errors "+
			"when trying to remediate.\n\n"+
			"# Remediations\n\n  ### a\n     -- fixed 1\n\n"+
			"# Non-remediated breaches\n\n", buf.String())
	})

	t.Run("noBreachRemediated", func(t *testing.T) {
		RunResultList = config.ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		RunResultList.TotalBreaches = 1
		RunResultList.TotalRemediations = 0
		RunResultList.Results = append(RunResultList.Results, config.Result{Name: "a", Status: config.Fail})
		buf = bytes.Buffer{}
		SimpleDisplay(w)
		assert.Equal("Breaches were detected but not all of them could be "+
			"fixed as they are either not supported yet or there were errors "+
			"when trying to remediate.\n\n"+
			"# Remediations\n\n"+
			"# Non-remediated breaches\n\n", buf.String())
	})
}
func TestJUnit(t *testing.T) {
	assert := assert.New(t)

	RunResultList = config.NewResultList(false)
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0"></testsuites>
`, buf.String())

	RunConfig.Checks = config.CheckMap{file.File: []config.Check{&file.FileCheck{
		CheckBase: config.CheckBase{Name: "a"},
	}}}
	RunResultList.Results = append(RunResultList.Results, config.Result{Name: "a", Status: config.Pass})
	buf = bytes.Buffer{}
	JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="file" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
    </testsuite>
</testsuites>
`, buf.String())

	RunConfig.Checks[file.File] = append(RunConfig.Checks[file.File], &file.FileCheck{
		CheckBase: config.CheckBase{Name: "b"},
	})
	RunResultList.Results = append(RunResultList.Results, config.Result{
		Name:     "b",
		Status:   config.Fail,
		Failures: []string{"Fail b"}})
	buf = bytes.Buffer{}
	JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="file" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
        <testcase name="b" classname="b">
            <error message="Fail b"></error>
        </testcase>
    </testsuite>
</testsuites>
`, buf.String())
}
