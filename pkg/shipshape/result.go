package shipshape

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
}
