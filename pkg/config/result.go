package config

import "sort"

// Sort reorders the Passes & Failures in order to get consistent output.
func (r *Result) Sort() {
	if len(r.Failures) > 0 {
		sort.Slice(r.Failures, func(i int, j int) bool {
			return r.Failures[i] < r.Failures[j]
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
