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
	Name         string   `json:"name"`
	Severity     string   `json:"severity"`
	CheckType    string   `json:"check-type"`
	Status       Status   `json:"status"`
	Passes       []string `json:"passes"`
	Breaches     []Breach `json:"breaches"`
	Warnings     []string `json:"warnings"`
	Remediations []string `json:"remediations"`
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
