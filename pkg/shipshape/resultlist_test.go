package shipshape_test

import (
	"sync"
	"testing"

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
