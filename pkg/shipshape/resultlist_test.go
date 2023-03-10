package shipshape_test

import (
	"bufio"
	"bytes"
	"sync"
	"testing"
	"text/tabwriter"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/file"
	. "github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/yaml"
	"github.com/stretchr/testify/assert"
)

func TestNewResultList(t *testing.T) {
	assert := assert.New(t)

	t.Run("emptyInit", func(t *testing.T) {
		rl := NewResultList(false)
		assert.Equal(rl.RemediationPerformed, false)
		assert.Equal(rl.Results, []config.Result{})
		assert.Equal(rl.CheckCountByType, map[config.CheckType]int{})
		assert.Equal(rl.BreachCountByType, map[config.CheckType]int{})
		assert.Equal(rl.BreachCountBySeverity, map[config.Severity]int{})
		assert.Equal(rl.RemediationCountByType, map[config.CheckType]int{})
	})

	t.Run("remediation", func(t *testing.T) {
		rl := NewResultList(true)
		assert.Equal(rl.RemediationPerformed, true)
		assert.Equal(rl.Results, []config.Result{})
		assert.Equal(rl.CheckCountByType, map[config.CheckType]int{})
		assert.Equal(rl.BreachCountByType, map[config.CheckType]int{})
		assert.Equal(rl.BreachCountBySeverity, map[config.Severity]int{})
		assert.Equal(rl.RemediationCountByType, map[config.CheckType]int{})
	})

}

func TestResultListStatus(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []config.Result{
			{Status: config.Pass},
			{Status: config.Pass},
			{Status: config.Pass},
		},
	}
	assert.Equal(config.Pass, rl.Status())

	rl.Results[0].Status = config.Fail
	assert.Equal(config.Fail, rl.Status())
}

func TestResultListIncrChecks(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		TotalChecks:      0,
		CheckCountByType: map[config.CheckType]int{},
	}
	rl.IncrChecks(file.File, 5)
	assert.Equal(5, int(rl.TotalChecks))
	assert.Equal(5, rl.CheckCountByType[file.File])

	rl.IncrChecks(yaml.Yaml, 5)
	assert.Equal(10, int(rl.TotalChecks))
	assert.Equal(5, rl.CheckCountByType[file.File])
	assert.Equal(5, rl.CheckCountByType[yaml.Yaml])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.IncrChecks(file.File, 1)
			rl.IncrChecks(yaml.Yaml, 1)
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalChecks))
	assert.Equal(105, rl.CheckCountByType[file.File])
	assert.Equal(105, rl.CheckCountByType[yaml.Yaml])
}

func TestResultListAddResult(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		TotalBreaches:         0,
		BreachCountByType:     map[config.CheckType]int{},
		BreachCountBySeverity: map[config.Severity]int{},

		TotalRemediations:      0,
		RemediationCountByType: map[config.CheckType]int{},
	}
	rl.AddResult(config.Result{
		Severity:     config.HighSeverity,
		CheckType:    file.File,
		Failures:     []string{"fail1", "fail2", "fail3", "fail4", "fail5"},
		Remediations: []string{"fixed1"},
	})
	assert.Equal(5, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[file.File])
	assert.Equal(5, rl.BreachCountBySeverity[config.HighSeverity])
	assert.Equal(1, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[file.File])

	rl.AddResult(config.Result{
		Severity:  config.CriticalSeverity,
		CheckType: yaml.Yaml,
		Failures:  []string{"fail1", "fail2", "fail3", "fail4", "fail5"},
	})
	assert.Equal(10, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[file.File])
	assert.Equal(5, rl.BreachCountByType[yaml.Yaml])
	assert.Equal(5, rl.BreachCountBySeverity[config.HighSeverity])
	assert.Equal(5, rl.BreachCountBySeverity[config.CriticalSeverity])
	assert.Equal(1, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[file.File])
	assert.Equal(0, rl.RemediationCountByType[yaml.Yaml])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.AddResult(config.Result{
				Severity:  config.HighSeverity,
				CheckType: file.File,
				Failures:  []string{"fail6"},
			})
			rl.AddResult(config.Result{
				Severity:     config.CriticalSeverity,
				CheckType:    yaml.Yaml,
				Failures:     []string{"fail6"},
				Remediations: []string{"fixed2", "fixed3"},
			})
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalBreaches))
	assert.Equal(105, rl.BreachCountByType[file.File])
	assert.Equal(105, rl.BreachCountByType[yaml.Yaml])
	assert.Equal(105, rl.BreachCountBySeverity[config.HighSeverity])
	assert.Equal(105, rl.BreachCountBySeverity[config.CriticalSeverity])
	assert.Equal(201, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[file.File])
	assert.Equal(200, rl.RemediationCountByType[yaml.Yaml])
}

func TestResultListGetBreachesByCheckName(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []config.Result{
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
		Results: []config.Result{
			{
				Severity: config.HighSeverity,
				Failures: []string{"failure1", "failure 2"},
			},
			{
				Severity: config.NormalSeverity,
				Failures: []string{"failure3", "failure 4"},
			},
		},
	}
	assert.EqualValues(
		[]string{"failure1", "failure 2"},
		rl.GetBreachesBySeverity(config.HighSeverity))
	assert.EqualValues(
		[]string{"failure3", "failure 4"},
		rl.GetBreachesBySeverity(config.NormalSeverity))
}

func TestResultListGetRemediationsByCheckName(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []config.Result{
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
		Results: []config.Result{
			{Name: "zcheck"},
			{Name: "hcheck"},
			{Name: "fcheck"},
			{Name: "acheck"},
			{Name: "ccheck"},
		},
	}
	rl.Sort()
	assert.EqualValues([]config.Result{
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
	rl = ResultList{Results: []config.Result{{Name: "a", Status: config.Pass}}}
	rl.TableDisplay(w)
	assert.Equal("NAME   STATUS   PASSES   FAILS\n"+
		"a      Pass              \n", buf.String())

	buf = bytes.Buffer{}
	rl = ResultList{
		Results: []config.Result{
			{Name: "a", Status: config.Pass},
			{Name: "b", Status: config.Pass},
			{Name: "c", Status: config.Pass},
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
		rl := NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.SimpleDisplay(w)
		assert.Equal("No result available; ensure your shipshape.yml is configured correctly.\n", buf.String())
	})

	t.Run("topShape", func(t *testing.T) {
		rl := NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.Results = append(rl.Results, config.Result{Name: "a", Status: config.Pass})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("breachesDetected", func(t *testing.T) {
		rl := NewResultList(false)
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.Results = append(rl.Results, config.Result{
			Name:     "b",
			Status:   config.Fail,
			Failures: []string{"Fail b"}})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("# Breaches were detected\n\n  ### b\n     -- Fail b\n\n", buf.String())
	})

	t.Run("topShapeRemediating", func(t *testing.T) {
		rl := ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.Results = append(rl.Results, config.Result{Name: "a", Status: config.Pass})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("Ship is in top shape; no breach detected!\n", buf.String())
	})

	t.Run("allBreachesRemediated", func(t *testing.T) {
		rl := ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.TotalRemediations = 1
		rl.Results = append(rl.Results, config.Result{Name: "a", Status: config.Pass, Remediations: []string{"fixed 1"}})
		buf = bytes.Buffer{}
		rl.SimpleDisplay(w)
		assert.Equal("Breaches were detected but were all fixed successfully!\n\n"+
			"  ### a\n     -- fixed 1\n\n", buf.String())
	})

	t.Run("someBreachesRemediated", func(t *testing.T) {
		rl := ResultList{RemediationPerformed: true}
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		rl.TotalRemediations = 1
		rl.TotalBreaches = 1
		rl.Results = append(rl.Results, config.Result{Name: "a", Status: config.Fail, Remediations: []string{"fixed 1"}})
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
		rl.Results = append(rl.Results, config.Result{Name: "a", Status: config.Fail})
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

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	RunResultList.JUnit(w)
	assert.Equal(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0"></testsuites>
`, buf.String())

	RunConfig.Checks = config.CheckMap{file.File: []config.Check{&file.FileCheck{
		CheckBase: config.CheckBase{Name: "a"},
	}}}
	RunResultList.Results = append(RunResultList.Results, config.Result{Name: "a", Status: config.Pass})
	buf = bytes.Buffer{}
	RunResultList.JUnit(w)
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
	RunResultList.JUnit(w)
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
