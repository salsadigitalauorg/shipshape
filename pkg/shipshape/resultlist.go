package shipshape

import (
	"sort"
	"sync"
	"sync/atomic"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

// Use locks to make map mutations concurrency-safe.
var lock = sync.RWMutex{}

func NewResultList(remediate bool) ResultList {
	return ResultList{
		RemediationPerformed:   remediate,
		Results:                []config.Result{},
		CheckCountByType:       map[config.CheckType]int{},
		BreachCountByType:      map[config.CheckType]int{},
		BreachCountBySeverity:  map[config.Severity]int{},
		RemediationCountByType: map[config.CheckType]int{},
	}
}

// Status calculates and returns the overall result of all check results.
func (rl *ResultList) Status() config.CheckStatus {
	for _, r := range rl.Results {
		if r.Status == config.Fail {
			return config.Fail
		}
	}
	return config.Pass
}

// IncrChecks increments the total checks count & checks count by type.
func (rl *ResultList) IncrChecks(ct config.CheckType, incr int) {
	atomic.AddUint32(&rl.TotalChecks, uint32(incr))

	lock.Lock()
	defer lock.Unlock()
	rl.CheckCountByType[ct] = rl.CheckCountByType[ct] + incr
}

// AddResult safely appends a check's result to the list.
func (rl *ResultList) AddResult(r config.Result) {
	lock.Lock()
	defer lock.Unlock()
	rl.Results = append(rl.Results, r)

	breachesIncr := len(r.Failures)
	atomic.AddUint32(&rl.TotalBreaches, uint32(breachesIncr))
	rl.BreachCountByType[r.CheckType] = rl.BreachCountByType[r.CheckType] + breachesIncr
	rl.BreachCountBySeverity[r.Severity] = rl.BreachCountBySeverity[r.Severity] + breachesIncr

	remediationsIncr := len(r.Remediations)
	atomic.AddUint32(&rl.TotalRemediations, uint32(remediationsIncr))
	rl.RemediationCountByType[r.CheckType] = rl.RemediationCountByType[r.CheckType] + remediationsIncr
}

// GetBreachesByCheckName fetches the list of failures by check name.
func (rl *ResultList) GetBreachesByCheckName(cn string) []string {
	var breaches []string
	for _, r := range rl.Results {
		if r.Name == cn {
			breaches = append(breaches, r.Failures...)
		}
	}
	return breaches
}

// GetBreachesBySeverity fetches the list of failures by severity.
func (rl *ResultList) GetBreachesBySeverity(s config.Severity) []string {
	var breaches []string

	for _, r := range rl.Results {
		if r.Severity == s {
			breaches = append(breaches, r.Failures...)
		}
	}
	return breaches
}

// GetBreachesByCheckName fetches the list of failures by check name.
func (rl *ResultList) GetRemediationsByCheckName(cn string) []string {
	var remediations []string
	for _, r := range rl.Results {
		if r.Name == cn {
			remediations = append(remediations, r.Remediations...)
		}
	}
	return remediations
}

// Sort reorders the results by name.
func (rl *ResultList) Sort() {
	sort.Slice(rl.Results, func(i int, j int) bool {
		return rl.Results[i].Name < rl.Results[j].Name
	})
}
