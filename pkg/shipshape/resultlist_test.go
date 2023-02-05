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

func TestResultListAddResult(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		TotalBreaches:         0,
		BreachCountByType:     map[CheckType]int{},
		BreachCountBySeverity: map[Severity]int{},

		TotalRemediations:      0,
		RemediationCountByType: map[CheckType]int{},
	}
	rl.AddResult(Result{
		Severity:     HighSeverity,
		CheckType:    File,
		Failures:     []string{"fail1", "fail2", "fail3", "fail4", "fail5"},
		Remediations: []string{"fixed1"},
	})
	assert.Equal(5, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[File])
	assert.Equal(5, rl.BreachCountBySeverity[HighSeverity])
	assert.Equal(1, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[File])

	rl.AddResult(Result{
		Severity:  CriticalSeverity,
		CheckType: Yaml,
		Failures:  []string{"fail1", "fail2", "fail3", "fail4", "fail5"},
	})
	assert.Equal(10, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[File])
	assert.Equal(5, rl.BreachCountByType[Yaml])
	assert.Equal(5, rl.BreachCountBySeverity[HighSeverity])
	assert.Equal(5, rl.BreachCountBySeverity[CriticalSeverity])
	assert.Equal(1, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[File])
	assert.Equal(0, rl.RemediationCountByType[Yaml])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.AddResult(Result{
				Severity:  HighSeverity,
				CheckType: File,
				Failures:  []string{"fail6"},
			})
			rl.AddResult(Result{
				Severity:     CriticalSeverity,
				CheckType:    Yaml,
				Failures:     []string{"fail6"},
				Remediations: []string{"fixed2", "fixed3"},
			})
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalBreaches))
	assert.Equal(105, rl.BreachCountByType[File])
	assert.Equal(105, rl.BreachCountByType[Yaml])
	assert.Equal(105, rl.BreachCountBySeverity[HighSeverity])
	assert.Equal(105, rl.BreachCountBySeverity[CriticalSeverity])
	assert.Equal(201, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[File])
	assert.Equal(200, rl.RemediationCountByType[Yaml])
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

func TestResultListGetRemediationsByCheckName(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []Result{
			{
				Name:         "check1",
				Remediations: []string{"fix1", "fix 2"},
			},
			{
				Name:         "check2",
				Remediations: []string{"fix3", "fix 4"},
			},
		},
	}
	assert.EqualValues(
		[]string{"fix1", "fix 2"},
		rl.GetRemediationsByCheckName("check1"))
	assert.EqualValues(
		[]string{"fix3", "fix 4"},
		rl.GetRemediationsByCheckName("check2"))
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

	t.Run("noResult", func(t *testing.T) {
		rl := NewResultList(&Config{})
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.SimpleDisplay(w)
		assert.Equal("No result available; ensure your shipshape.yml is configured correctly.\n", buf.String())
	})

	t.Run("topShape", func(t *testing.T) {
		rl := NewResultList(&Config{})
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.Results = append(rl.Results, Result{Name: "a", Status: Pass})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("breachesDetected", func(t *testing.T) {
		rl := NewResultList(&Config{})
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.Results = append(rl.Results, Result{
			Name:     "b",
			Status:   Fail,
			Failures: []string{"Fail b"}})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("# Breaches were detected\n\n  ### b\n     -- Fail b\n\n", buf.String())
	})

	t.Run("topShapeRemediating", func(t *testing.T) {
		rl := ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.Results = append(rl.Results, Result{Name: "a", Status: Pass})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("allBreachesRemediated", func(t *testing.T) {
		rl := ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.TotalRemediations = 1
		rl.Results = append(rl.Results, Result{Name: "a", Status: Pass, Remediations: []string{"fixed 1"}})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("Breaches were detected but were all fixed successfully!\n\n", buf.String())
	})

	t.Run("someBreachesRemediated", func(t *testing.T) {
		rl := ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.TotalRemediations = 1
		rl.TotalBreaches = 1
		rl.Results = append(rl.Results, Result{Name: "a", Status: Fail, Remediations: []string{"fixed 1"}})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("Breaches were detected but not all of them could be "+
			"fixed as they are either not supported yet or there were errors "+
			"when trying to remediate.\n\n"+
			"# Remediations\n\n  ### a\n     -- fixed 1\n\n"+
			"# Non-remediated breaches\n\n", buf.String())
	})

	t.Run("noBreachRemediated", func(t *testing.T) {
		rl := ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.TotalBreaches = 1
		rl.TotalRemediations = 0
		rl.Results = append(rl.Results, Result{Name: "a", Status: Fail})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("Breaches were detected but not all of them could be "+
			"fixed as they are either not supported yet or there were errors "+
			"when trying to remediate.\n\n"+
			"# Remediations\n\n"+
			"# Non-remediated breaches\n\n", buf.String())
	})
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
