package result_test

import (
	"sync"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
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

const testCheckType config.CheckType = "test-check"
const testCheck2Type config.CheckType = "test-check2"

func TestResultListIncrChecks(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		TotalChecks:      0,
		CheckCountByType: map[string]int{},
	}
	rl.IncrChecks(string(testCheckType), 5)
	assert.Equal(5, int(rl.TotalChecks))
	assert.Equal(5, rl.CheckCountByType[string(testCheckType)])

	rl.IncrChecks(string(testCheck2Type), 5)
	assert.Equal(10, int(rl.TotalChecks))
	assert.Equal(5, rl.CheckCountByType[string(testCheckType)])
	assert.Equal(5, rl.CheckCountByType[string(testCheck2Type)])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.IncrChecks(string(testCheckType), 1)
			rl.IncrChecks(string(testCheck2Type), 1)
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalChecks))
	assert.Equal(105, rl.CheckCountByType[string(testCheckType)])
	assert.Equal(105, rl.CheckCountByType[string(testCheck2Type)])
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
		Severity:  "high",
		CheckType: string(testCheckType),
		Breaches: []Breach{
			&ValueBreach{Value: "fail1"},
			&ValueBreach{Value: "fail2"},
			&ValueBreach{Value: "fail3"},
			&ValueBreach{Value: "fail4"},
			&ValueBreach{Value: "fail5"},
		},
		Remediations: []string{"fixed1"},
	})
	assert.Equal(5, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[string(testCheckType)])
	assert.Equal(5, rl.BreachCountBySeverity["high"])
	assert.Equal(1, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[string(testCheckType)])

	rl.AddResult(Result{
		Severity:  "critical",
		CheckType: string(testCheck2Type),
		Breaches: []Breach{
			&ValueBreach{Value: "fail1"},
			&ValueBreach{Value: "fail2"},
			&ValueBreach{Value: "fail3"},
			&ValueBreach{Value: "fail4"},
			&ValueBreach{Value: "fail5"},
		},
	})
	assert.Equal(10, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[string(testCheckType)])
	assert.Equal(5, rl.BreachCountByType[string(testCheck2Type)])
	assert.Equal(5, rl.BreachCountBySeverity["high"])
	assert.Equal(5, rl.BreachCountBySeverity["critical"])
	assert.Equal(1, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[string(testCheckType)])
	assert.Equal(0, rl.RemediationCountByType[string(testCheck2Type)])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.AddResult(Result{
				Severity:  "high",
				CheckType: string(testCheckType),
				Breaches:  []Breach{&ValueBreach{Value: "fail6"}},
			})
			rl.AddResult(Result{
				Severity:     "critical",
				CheckType:    string(testCheck2Type),
				Breaches:     []Breach{&ValueBreach{Value: "fail7"}},
				Remediations: []string{"fixed2", "fixed3"},
			})
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalBreaches))
	assert.Equal(105, rl.BreachCountByType[string(testCheckType)])
	assert.Equal(105, rl.BreachCountByType[string(testCheck2Type)])
	assert.Equal(105, rl.BreachCountBySeverity["high"])
	assert.Equal(105, rl.BreachCountBySeverity["critical"])
	assert.Equal(201, int(rl.TotalRemediations))
	assert.Equal(1, rl.RemediationCountByType[string(testCheckType)])
	assert.Equal(200, rl.RemediationCountByType[string(testCheck2Type)])
}

func TestResultListGetBreachesByCheckName(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []Result{
			{
				Name: "check1",
				Breaches: []Breach{
					&ValueBreach{Value: "failure1"},
					&ValueBreach{Value: "failure 2"},
				},
			},
			{
				Name: "check2",
				Breaches: []Breach{
					&ValueBreach{Value: "failure3"},
					&ValueBreach{Value: "failure 4"},
				},
			},
		},
	}
	assert.EqualValues(
		[]Breach{
			&ValueBreach{Value: "failure1"},
			&ValueBreach{Value: "failure 2"},
		},
		rl.GetBreachesByCheckName("check1"))
	assert.EqualValues(
		[]Breach{
			&ValueBreach{Value: "failure3"},
			&ValueBreach{Value: "failure 4"},
		},
		rl.GetBreachesByCheckName("check2"))
}

func TestResultListGetBreachesBySeverity(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []Result{
			{
				Severity: "high",
				Breaches: []Breach{
					&ValueBreach{Value: "failure1"},
					&ValueBreach{Value: "failure 2"},
				},
			},
			{
				Severity: "normal",
				Breaches: []Breach{
					&ValueBreach{Value: "failure3"},
					&ValueBreach{Value: "failure 4"},
				},
			},
		},
	}
	assert.EqualValues(
		[]Breach{
			&ValueBreach{Value: "failure1"},
			&ValueBreach{Value: "failure 2"},
		},
		rl.GetBreachesBySeverity("high"))
	assert.EqualValues(
		[]Breach{
			&ValueBreach{Value: "failure3"},
			&ValueBreach{Value: "failure 4"},
		},
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
