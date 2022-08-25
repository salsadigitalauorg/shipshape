package shipshape_test

import (
	"bufio"
	"bytes"
	"sync"
	"testing"
	"text/tabwriter"

	. "github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestResultListStatus(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []Result{
			{Status: Pass},
			{Status: Pass},
			{Status: Pass},
		},
	}
	assert.Equal(Pass, rl.Status())

	rl.Results[0].Status = Fail
	assert.Equal(Fail, rl.Status())
}

func TestResultListIncrChecks(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		TotalChecks:      0,
		CheckCountByType: map[CheckType]int{},
	}
	rl.IncrChecks(File, 5)
	assert.Equal(5, int(rl.TotalChecks))
	assert.Equal(5, rl.CheckCountByType[File])

	rl.IncrChecks(Yaml, 5)
	assert.Equal(10, int(rl.TotalChecks))
	assert.Equal(5, rl.CheckCountByType[File])
	assert.Equal(5, rl.CheckCountByType[Yaml])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.IncrChecks(File, 1)
			rl.IncrChecks(Yaml, 1)
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalChecks))
	assert.Equal(105, rl.CheckCountByType[File])
	assert.Equal(105, rl.CheckCountByType[Yaml])
}

func TestResultListIncrBreaches(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		TotalBreaches:         0,
		BreachCountByType:     map[CheckType]int{},
		BreachCountBySeverity: map[Severity]int{},
	}
	rl.IncrBreaches(File, HighSeverity, 5)
	assert.Equal(5, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[File])
	assert.Equal(5, rl.BreachCountBySeverity[HighSeverity])

	rl.IncrBreaches(Yaml, CriticalSeverity, 5)
	assert.Equal(10, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[File])
	assert.Equal(5, rl.BreachCountByType[Yaml])
	assert.Equal(5, rl.BreachCountBySeverity[HighSeverity])
	assert.Equal(5, rl.BreachCountBySeverity[CriticalSeverity])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.IncrBreaches(File, HighSeverity, 1)
			rl.IncrBreaches(Yaml, CriticalSeverity, 1)
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalBreaches))
	assert.Equal(105, rl.BreachCountByType[File])
	assert.Equal(105, rl.BreachCountByType[Yaml])
	assert.Equal(105, rl.BreachCountBySeverity[HighSeverity])
	assert.Equal(105, rl.BreachCountBySeverity[CriticalSeverity])
}

func TestResultListGetBreachesByCheckName(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []Result{
			{
				Name:     "check1",
				Failures: []string{"failure1", "failure 2"},
			},
			{
				Name:     "check2",
				Failures: []string{"failure3", "failure 4"},
			},
		},
	}
	assert.EqualValues(
		[]string{"failure1", "failure 2"},
		rl.GetBreachesByCheckName("check1"))
	assert.EqualValues(
		[]string{"failure3", "failure 4"},
		rl.GetBreachesByCheckName("check2"))
}

func TestResultListGetBreachesBySeverity(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []Result{
			{
				Severity: HighSeverity,
				Failures: []string{"failure1", "failure 2"},
			},
			{
				Severity: NormalSeverity,
				Failures: []string{"failure3", "failure 4"},
			},
		},
	}
	assert.EqualValues(
		[]string{"failure1", "failure 2"},
		rl.GetBreachesBySeverity(HighSeverity))
	assert.EqualValues(
		[]string{"failure3", "failure 4"},
		rl.GetBreachesBySeverity(NormalSeverity))
}

func TestResultListSort(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []Result{
			{Name: "zcheck"},
			{Name: "hcheck"},
			{Name: "fcheck"},
			{Name: "acheck"},
			{Name: "ccheck"},
		},
	}
	rl.Sort()
	assert.EqualValues([]Result{
		{Name: "acheck"},
		{Name: "ccheck"},
		{Name: "fcheck"},
		{Name: "hcheck"},
		{Name: "zcheck"},
	}, rl.Results)
}

func TestResultListTableDisplay(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	rl := ResultList{}
	rl.TableDisplay(w)
	assert.Equal(
		"No result available; ensure your shipshape.yml is configured correctly.\n",
		buf.String())

	buf = bytes.Buffer{}
	rl = ResultList{Results: []Result{{Name: "a", Status: Pass}}}
	rl.TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n", buf.String())

	buf = bytes.Buffer{}
	rl = ResultList{
		Results: []Result{
			{Name: "a", Status: Pass},
			{Name: "b", Status: Pass},
			{Name: "c", Status: Pass},
		},
	}
	rl.TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n"+
		"b      Pass              \n"+
		"c      Pass              \n",
		buf.String())

	buf = bytes.Buffer{}
	rl = ResultList{
		Results: []Result{
			{
				Name:   "a",
				Status: Pass,
				Passes: []string{"Pass a", "Pass ab"},
			},
			{
				Name:   "b",
				Status: Pass,
				Passes: []string{"Pass b", "Pass bb", "Pass bc"},
			},
			{
				Name:     "c",
				Status:   Fail,
				Failures: []string{"Fail c", "Fail cb"},
			},
			{
				Name:     "d",
				Status:   Fail,
				Passes:   []string{"Pass d", "Pass db"},
				Failures: []string{"Fail c", "Fail cb"},
			},
		},
	}
	rl.TableDisplay(w)
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

func TestResultListSimpleDisplay(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	rl.SimpleDisplay(w)
	assert.Equal("No result available; ensure your shipshape.yml is configured correctly.\n", buf.String())

	rl.Results = append(rl.Results, Result{Name: "a", Status: Pass})
	buf = bytes.Buffer{}
	rl.SimpleDisplay(w)
	assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())

	rl.Results = append(rl.Results, Result{
		Name:     "b",
		Status:   Fail,
		Failures: []string{"Fail b"}})
	buf = bytes.Buffer{}
	rl.SimpleDisplay(w)
	assert.Equal("Breaches were detected!\n\n### b\n   -- Fail b\n\n", buf.String())
}
func TestResultListJUnit(t *testing.T) {
	assert := assert.New(t)

	cfg := &Config{}
	rl := NewResultList(cfg)
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	rl.JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0"></testsuites>
`, buf.String())

	cfg.Checks = CheckMap{File: []Check{&FileCheck{
		CheckBase: CheckBase{Name: "a"},
	}}}
	rl.Results = append(rl.Results, Result{Name: "a", Status: Pass})
	buf = bytes.Buffer{}
	rl.JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="file" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
    </testsuite>
</testsuites>
`, buf.String())

	cfg.Checks[File] = append(cfg.Checks[File], &FileCheck{
		CheckBase: CheckBase{Name: "b"},
	})
	rl.Results = append(rl.Results, Result{
		Name:     "b",
		Status:   Fail,
		Failures: []string{"Fail b"}})
	buf = bytes.Buffer{}
	rl.JUnit(w)
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
