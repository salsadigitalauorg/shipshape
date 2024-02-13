package result

import (
	"sort"
	"sync"
	"sync/atomic"
)

// ResultList is a wrapper around a list of results, providing some useful
// methods to manipulate and use it.
type ResultList struct {
	RemediationPerformed  bool              `json:"remediation-performed"`
	TotalChecks           uint32            `json:"total-checks"`
	TotalBreaches         uint32            `json:"total-breaches"`
	RemediationTotals     map[string]uint32 `json:"remediation-totals"`
	CheckCountByType      map[string]int    `json:"check-count-by-type"`
	BreachCountByType     map[string]int    `json:"breach-count-by-type"`
	BreachCountBySeverity map[string]int    `json:"breach-count-by-severity"`
	Results               []Result          `json:"results"`
}

// Use locks to make map mutations concurrency-safe.
var lock = sync.RWMutex{}

func NewResultList(remediate bool) ResultList {
	rl := ResultList{
		RemediationPerformed:  remediate,
		Results:               []Result{},
		CheckCountByType:      map[string]int{},
		BreachCountByType:     map[string]int{},
		BreachCountBySeverity: map[string]int{},
	}
	return rl
}

// IncrChecks increments the total checks count & checks count by type.
func (rl *ResultList) IncrChecks(ct string, incr int) {
	atomic.AddUint32(&rl.TotalChecks, uint32(incr))

	lock.Lock()
	defer lock.Unlock()
	rl.CheckCountByType[ct] = rl.CheckCountByType[ct] + incr
}

// AddResult safely appends a check's result to the list.
func (rl *ResultList) AddResult(r Result) {
	lock.Lock()
	defer lock.Unlock()
	rl.Results = append(rl.Results, r)

	breachesIncr := len(r.Breaches)
	atomic.AddUint32(&rl.TotalBreaches, uint32(breachesIncr))
	rl.BreachCountByType[r.CheckType] = rl.BreachCountByType[r.CheckType] + breachesIncr
	rl.BreachCountBySeverity[r.Severity] = rl.BreachCountBySeverity[r.Severity] + breachesIncr
}

// Status calculates and returns the overall result of all check results.
func (rl *ResultList) Status() Status {
	for _, r := range rl.Results {
		if r.Status == Fail {
			return Fail
		}
	}
	return Pass
}

// RemediationTotalsCount calculates the total number of unsupported,
// successful, failed and partial remediations across all checks.
func (rl *ResultList) RemediationTotalsCount() {
	rl.RemediationTotals = map[string]uint32{
		"unsupported": 0,
		"successful":  0,
		"failed":      0,
		"partial":     0,
	}
	for _, r := range rl.Results {
		unsupported, successful, failed, partial := r.RemediationsCount()
		rl.RemediationTotals["unsupported"] = rl.RemediationTotals["unsupported"] + unsupported
		rl.RemediationTotals["successful"] = rl.RemediationTotals["successful"] + successful
		rl.RemediationTotals["failed"] = rl.RemediationTotals["failed"] + failed
		rl.RemediationTotals["partial"] = rl.RemediationTotals["partial"] + partial
	}
}

// RemediationStatus calculates and returns the overall result of
// remediation for all breaches.
func (rl *ResultList) RemediationStatus() RemediationStatus {
	if !rl.RemediationPerformed {
		return ""
	}

	if rl.RemediationTotals["partial"] > 0 ||
		(rl.RemediationTotals["successful"] > 0 &&
			(rl.RemediationTotals["failed"] > 0 ||
				rl.RemediationTotals["unsupported"] > 0)) {
		return RemediationStatusPartial
	}
	if rl.RemediationTotals["unsupported"] > 0 &&
		rl.RemediationTotals["successful"] == 0 &&
		rl.RemediationTotals["failed"] == 0 &&
		rl.RemediationTotals["partial"] == 0 {
		return RemediationStatusNoSupport
	}
	if rl.RemediationTotals["failed"] > 0 &&
		rl.RemediationTotals["successful"] == 0 &&
		rl.RemediationTotals["unsupported"] == 0 &&
		rl.RemediationTotals["partial"] == 0 {
		return RemediationStatusFailed
	}
	return RemediationStatusSuccess
}

// GetBreachesByCheckName fetches the list of failures by check name.
func (rl *ResultList) GetBreachesByCheckName(cn string) []Breach {
	var breaches []Breach
	for _, r := range rl.Results {
		if r.Name == cn {
			breaches = append(breaches, r.Breaches...)
		}
	}
	return breaches
}

// GetBreachesBySeverity fetches the list of failures by severity.
func (rl *ResultList) GetBreachesBySeverity(s string) []Breach {
	var breaches []Breach

	for _, r := range rl.Results {
		if r.Severity == s {
			breaches = append(breaches, r.Breaches...)
		}
	}
	return breaches
}

// Sort reorders the results by name.
func (rl *ResultList) Sort() {
	sort.Slice(rl.Results, func(i int, j int) bool {
		return rl.Results[i].Name < rl.Results[j].Name
	})
}
