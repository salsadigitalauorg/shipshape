package result

import (
	"sort"
)

type Status string

const (
	Pass Status = "Pass"
	Fail Status = "Fail"
)

// Result provides the structure for a Check's outcome.
type Result struct {
	Name              string            `json:"name"`
	Severity          string            `json:"severity"`
	CheckType         string            `json:"check-type"`
	Passes            []string          `json:"passes"`
	Breaches          []Breach          `json:"breaches"`
	Warnings          []string          `json:"warnings"`
	Status            Status            `json:"status"`
	RemediationStatus RemediationStatus `json:"remediation-status"`
}

// Sort reorders the Passes & Failures in order to get consistent output.
func (r *Result) Sort() {
	if len(r.Breaches) > 0 {
		sort.Slice(r.Breaches, func(i int, j int) bool {
			return r.Breaches[i].GetCheckName() < r.Breaches[j].GetCheckName()
		})
	}

	if len(r.Passes) > 0 {
		sort.Slice(r.Passes, func(i int, j int) bool {
			return r.Passes[i] < r.Passes[j]
		})
	}

	if len(r.Warnings) > 0 {
		sort.Slice(r.Warnings, func(i int, j int) bool {
			return r.Warnings[i] < r.Warnings[j]
		})
	}
}

// RemediationsCount returns the number of unsupported, successful, failed and
// partial for all attempted remediations.
func (r *Result) RemediationsCount() (uint32, uint32, uint32, uint32) {
	unsupported := uint32(0)
	successful := uint32(0)
	failed := uint32(0)
	partial := uint32(0)
	for _, b := range r.Breaches {
		switch b.GetRemediation().Status {
		case RemediationStatusNoSupport:
			unsupported++
		case RemediationStatusSuccess:
			successful++
		case RemediationStatusFailed:
			failed++
		case RemediationStatusPartial:
			partial++
		}
	}
	return unsupported, successful, failed, partial
}

// DetermineResultStatus determines the overall status of the result based on
// the breaches and remediation status.
func (r *Result) DetermineResultStatus(remediationPerformed bool) {
	r.Sort()

	// Remediation status.
	if remediationPerformed {
		unsupported, success, failed, partial := r.RemediationsCount()
		if partial > 0 || (success > 0 && (failed > 0 || unsupported > 0)) {
			r.RemediationStatus = RemediationStatusPartial
			r.Status = Fail
			return
		}
		if unsupported > 0 && success == 0 && failed == 0 && partial == 0 {
			r.RemediationStatus = RemediationStatusNoSupport
			r.Status = Fail
			return
		}
		if failed > 0 && success == 0 && unsupported == 0 && partial == 0 {
			r.RemediationStatus = RemediationStatusFailed
			r.Status = Fail
			return
		}
		r.RemediationStatus = RemediationStatusSuccess
		r.Status = Pass
		return
	}

	// Overall status.
	if len(r.Breaches) > 0 {
		r.Status = Fail
		return
	}
	r.Status = Pass
}
