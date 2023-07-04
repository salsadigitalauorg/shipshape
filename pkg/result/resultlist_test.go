package result_test

import (
	"sync"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/checks/file"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	. "github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestNewResultList(t *testing.T) {
	assert := assert.New(t)

	t.Run("emptyInit", func(t *testing.T) {
		rl := NewResultList(false)
		assert.Equal(rl.RemediationPerformed, false)
		assert.Equal(rl.Results, []Result{})
		assert.Equal(rl.CheckCountByType, map[string]int{})
		assert.Equal(rl.BreachCountByType, map[string]int{})
		assert.Equal(rl.BreachCountBySeverity, map[string]int{})
		assert.Equal(rl.RemediationCountByType, map[string]int{})
	})

	t.Run("remediation", func(t *testing.T) {
		rl := NewResultList(true)
		assert.Equal(rl.RemediationPerformed, true)
		assert.Equal(rl.Results, []Result{})
		assert.Equal(rl.CheckCountByType, map[string]int{})
		assert.Equal(rl.BreachCountByType, map[string]int{})
		assert.Equal(rl.BreachCountBySeverity, map[string]int{})
		assert.Equal(rl.RemediationCountByType, map[string]int{})
	})

}

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
		CheckCountByType: map[string]int{},
	}
	rl.IncrChecks(string(file.File), 5)
	assert.Equal(5, int(rl.TotalChecks))
	assert.Equal(5, rl.CheckCountByType[string(file.File)])

	rl.IncrChecks(string(yaml.Yaml), 5)
	assert.Equal(10, int(rl.TotalChecks))
	assert.Equal(5, rl.CheckCountByType[string(file.File)])
	assert.Equal(5, rl.CheckCountByType[string(yaml.Yaml)])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.IncrChecks(string(file.File), 1)
			rl.IncrChecks(string(yaml.Yaml), 1)
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalChecks))
	assert.Equal(105, rl.CheckCountByType[string(file.File)])
	assert.Equal(105, rl.CheckCountByType[string(yaml.Yaml)])
}

func TestResultListAddResult(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		TotalBreaches:         0,
		BreachCountByType:     map[string]int{},
		BreachCountBySeverity: map[string]int{},

		TotalRemediations:      0,
		RemediationCountByType: map[string]int{},
	}
	rl.AddResult(Result{
		Severity:     "high",
		CheckType:    string(file.File),
		Failures:     []string{"fail1", "fail2", "fail3", "fail4", "fail5"},
		Remediations: []string{"fixed1"},
	})
	assert.Equal(5, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[string(file.File)])
	assert.Equal(5, rl.BreachCountBySeverity["high"])
	assert.Equal(1, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[string(file.File)])

	rl.AddResult(Result{
		Severity:  "critical",
		CheckType: string(yaml.Yaml),
		Failures:  []string{"fail1", "fail2", "fail3", "fail4", "fail5"},
	})
	assert.Equal(10, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[string(file.File)])
	assert.Equal(5, rl.BreachCountByType[string(yaml.Yaml)])
	assert.Equal(5, rl.BreachCountBySeverity["high"])
	assert.Equal(5, rl.BreachCountBySeverity["critical"])
	assert.Equal(1, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[string(file.File)])
	assert.Equal(0, rl.RemediationCountByType[string(yaml.Yaml)])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.AddResult(Result{
				Severity:  "high",
				CheckType: string(file.File),
				Failures:  []string{"fail6"},
			})
			rl.AddResult(Result{
				Severity:     "critical",
				CheckType:    string(yaml.Yaml),
				Failures:     []string{"fail6"},
				Remediations: []string{"fixed2", "fixed3"},
			})
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalBreaches))
	assert.Equal(105, rl.BreachCountByType[string(file.File)])
	assert.Equal(105, rl.BreachCountByType[string(yaml.Yaml)])
	assert.Equal(105, rl.BreachCountBySeverity["high"])
	assert.Equal(105, rl.BreachCountBySeverity["critical"])
	assert.Equal(201, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[string(file.File)])
	assert.Equal(200, rl.RemediationCountByType[string(yaml.Yaml)])
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
				Severity: "high",
				Failures: []string{"failure1", "failure 2"},
			},
			{
				Severity: "normal",
				Failures: []string{"failure3", "failure 4"},
			},
		},
	}
	assert.EqualValues(
		[]string{"failure1", "failure 2"},
		rl.GetBreachesBySeverity("high"))
	assert.EqualValues(
		[]string{"failure3", "failure 4"},
		rl.GetBreachesBySeverity("normal"))
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
