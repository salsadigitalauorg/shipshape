package result_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	. "github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestNewResultList(t *testing.T) {
	assert := assert.New(t)

	t.Run("emptyInit", func(t *testing.T) {
		rl := NewResultList(false)
		assert.Equal(false, rl.RemediationPerformed)
		assert.Equal([]Result{}, rl.Results)
		assert.Equal(map[string]int{}, rl.CheckCountByType)
		assert.Equal(map[string]int{}, rl.BreachCountByType)
		assert.Equal(map[string]int{}, rl.BreachCountBySeverity)
		assert.Nil(rl.RemediationTotals)
	})

	t.Run("remediation", func(t *testing.T) {
		rl := NewResultList(true)
		assert.Equal(true, rl.RemediationPerformed)
		assert.Equal([]Result{}, rl.Results)
		assert.Equal(map[string]int{}, rl.CheckCountByType)
		assert.Equal(map[string]int{}, rl.BreachCountByType)
		assert.Equal(map[string]int{}, rl.BreachCountBySeverity)
		assert.Nil(rl.RemediationTotals)
	})
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

		RemediationTotals: map[string]uint32{"successful": 0},
	}
	rl.AddResult(Result{
		Severity:  "high",
		CheckType: string(testCheckType),
		Breaches: []breach.Breach{
			&breach.ValueBreach{Value: "fail1", RemediationResult: breach.RemediationResult{
				Status:   breach.RemediationStatusSuccess,
				Messages: []string{"fixed1"},
			}},
			&breach.ValueBreach{Value: "fail2"},
			&breach.ValueBreach{Value: "fail3"},
			&breach.ValueBreach{Value: "fail4"},
			&breach.ValueBreach{Value: "fail5"},
		},
	})
	assert.Equal(5, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[string(testCheckType)])
	assert.Equal(5, rl.BreachCountBySeverity["high"])

	rl.AddResult(Result{
		Severity:  "critical",
		CheckType: string(testCheck2Type),
		Breaches: []breach.Breach{
			&breach.ValueBreach{Value: "fail1"},
			&breach.ValueBreach{Value: "fail2"},
			&breach.ValueBreach{Value: "fail3"},
			&breach.ValueBreach{Value: "fail4"},
			&breach.ValueBreach{Value: "fail5"},
		},
	})
	assert.Equal(10, int(rl.TotalBreaches))
	assert.Equal(5, rl.BreachCountByType[string(testCheckType)])
	assert.Equal(5, rl.BreachCountByType[string(testCheck2Type)])
	assert.Equal(5, rl.BreachCountBySeverity["high"])
	assert.Equal(5, rl.BreachCountBySeverity["critical"])

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.AddResult(Result{
				Severity:  "high",
				CheckType: string(testCheckType),
				Breaches:  []breach.Breach{&breach.ValueBreach{Value: "fail6"}},
			})
			rl.AddResult(Result{
				Severity:  "critical",
				CheckType: string(testCheck2Type),
				Breaches: []breach.Breach{&breach.ValueBreach{
					Value: "fail7",
					RemediationResult: breach.RemediationResult{
						Status:   breach.RemediationStatusSuccess,
						Messages: []string{"fixed2", "fixed3"},
					},
				}},
			})
		}()
	}
	wg.Wait()
	assert.Equal(210, int(rl.TotalBreaches))
	assert.Equal(105, rl.BreachCountByType[string(testCheckType)])
	assert.Equal(105, rl.BreachCountByType[string(testCheck2Type)])
	assert.Equal(105, rl.BreachCountBySeverity["high"])
	assert.Equal(105, rl.BreachCountBySeverity["critical"])
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

func TestResultListRemediationTotalsCount(t *testing.T) {
	tt := []struct {
		name     string
		results  []Result
		expected map[string]uint32
	}{
		{
			name: "allSuccess",
			results: []Result{
				{Breaches: []breach.Breach{
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}},
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}},
				}},
				{Breaches: []breach.Breach{
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}},
				}},
			},
			expected: map[string]uint32{
				"unsupported": 0,
				"successful":  3,
				"failed":      0,
				"partial":     0,
			},
		},
		{
			name: "countingWorks",
			results: []Result{
				{Breaches: []breach.Breach{
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}},
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}},
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed}},
				}},
				{Breaches: []breach.Breach{
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}},
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusPartial}},
				}},
				{Breaches: []breach.Breach{
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed}},
					&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusNoSupport}},
				}},
			},
			expected: map[string]uint32{
				"unsupported": 1,
				"successful":  3,
				"failed":      2,
				"partial":     1,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			rl := ResultList{
				Results: tc.results,
			}
			rl.RemediationTotalsCount()
			assert.Equal(tc.expected, rl.RemediationTotals)
		})
	}
}

func TestResultListRemediationStatus(t *testing.T) {
	tt := []struct {
		name                 string
		remediationPerformed bool
		results              []Result
		expected             breach.RemediationStatus
	}{
		{
			name:                 "noRemediation",
			remediationPerformed: false,
			results:              []Result{{Breaches: []breach.Breach{&breach.ValueBreach{}}}},
			expected:             "",
		},
		{
			name:                 "allSuccess",
			remediationPerformed: true,
			results: []Result{
				{
					Breaches: []breach.Breach{&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}}},
				},
				{
					Breaches: []breach.Breach{&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}}},
				},
				{
					Breaches: []breach.Breach{&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}}},
				},
			},
			expected: breach.RemediationStatusSuccess,
		},
		{
			name:                 "partial",
			remediationPerformed: true,
			results: []Result{
				{
					Breaches: []breach.Breach{&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusPartial}}},
				},
				{
					Breaches: []breach.Breach{&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed}}},
				},
			},
			expected: breach.RemediationStatusPartial,
		},
		{
			name:                 "fail",
			remediationPerformed: true,
			results: []Result{
				{
					Breaches: []breach.Breach{&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed}}},
				},
				{
					Breaches: []breach.Breach{&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed}}},
				},
			},
			expected: breach.RemediationStatusFailed,
		},
		{
			name:                 "unsupported",
			remediationPerformed: true,
			results: []Result{
				{
					Breaches: []breach.Breach{&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusNoSupport}}},
				},
				{
					Breaches: []breach.Breach{&breach.ValueBreach{RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusNoSupport}}},
				},
			},
			expected: breach.RemediationStatusNoSupport,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			rl := ResultList{
				RemediationPerformed: tc.remediationPerformed,
				Results:              tc.results,
			}
			rl.RemediationTotalsCount()
			assert.Equal(tc.expected, rl.RemediationStatus())
		})
	}
}

func TestResultListGetBreachesByCheckName(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []Result{
			{
				Name: "check1",
				Breaches: []breach.Breach{
					&breach.ValueBreach{Value: "failure1"},
					&breach.ValueBreach{Value: "failure 2"},
				},
			},
			{
				Name: "check2",
				Breaches: []breach.Breach{
					&breach.ValueBreach{Value: "failure3"},
					&breach.ValueBreach{Value: "failure 4"},
				},
			},
		},
	}
	assert.EqualValues(
		[]breach.Breach{
			&breach.ValueBreach{Value: "failure1"},
			&breach.ValueBreach{Value: "failure 2"},
		},
		rl.GetBreachesByCheckName("check1"))
	assert.EqualValues(
		[]breach.Breach{
			&breach.ValueBreach{Value: "failure3"},
			&breach.ValueBreach{Value: "failure 4"},
		},
		rl.GetBreachesByCheckName("check2"))
}

func TestResultListGetBreachesBySeverity(t *testing.T) {
	assert := assert.New(t)

	rl := ResultList{
		Results: []Result{
			{
				Severity: "high",
				Breaches: []breach.Breach{
					&breach.ValueBreach{Value: "failure1"},
					&breach.ValueBreach{Value: "failure 2"},
				},
			},
			{
				Severity: "normal",
				Breaches: []breach.Breach{
					&breach.ValueBreach{Value: "failure3"},
					&breach.ValueBreach{Value: "failure 4"},
				},
			},
		},
	}
	assert.EqualValues(
		[]breach.Breach{
			&breach.ValueBreach{Value: "failure1"},
			&breach.ValueBreach{Value: "failure 2"},
		},
		rl.GetBreachesBySeverity("high"))
	assert.EqualValues(
		[]breach.Breach{
			&breach.ValueBreach{Value: "failure3"},
			&breach.ValueBreach{Value: "failure 4"},
		},
		rl.GetBreachesBySeverity("normal"))
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
