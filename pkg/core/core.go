// Package core provides the basic types and base functions for running checks
// as well as the file check.
package core

import (
	"fmt"
	"text/tabwriter"
)

// Init acts as the constructor of a check and sets some initial values.
func (c *CheckBase) Init(pd string, ct CheckType) {
	c.ProjectDir = pd
	if c.Result.CheckType == "" {
		c.Result = Result{Name: c.Name, CheckType: ct}
	}
}

// GetName returns the name of a check.
func (c *CheckBase) GetName() string {
	return c.Name
}

// RequiresData indicates whether the check requires a DataMap to run against.
// It is designed as opt-out, so remember to set it to false if you are creating
// a check that does not require the DataMap.
func (c *CheckBase) RequiresData() bool { return true }

// FetchData contains the logic for fetching the data over which the check is
// going to run.
// This is where c.DataMap should be populated.
func (c *CheckBase) FetchData() {}

// HasData determines whether the dataMap has been populated or not.
// The Check can optionally be marked as failed if the dataMap is not populated.
func (c *CheckBase) HasData(failCheck bool) bool {
	if c.DataMap == nil {
		if failCheck {
			c.FailCheck("no data available")
		}
		return false
	}
	return true
}

// UnmarshalDataMap attempts to parse the DataMap into a structure that
// can be used to execute the check. Any failure here should fail the check.
func (c *CheckBase) UnmarshalDataMap() {}

// FailCheck sets the Check as Fail.
func (c *CheckBase) FailCheck(estr string) {
	c.Result.Status = Fail
	c.Result.Failures = append(
		c.Result.Failures,
		estr,
	)
}

// RunCheck contains the core logic for running the check and generating
// the result.
// This is where c.Result should be populated.
func (c *CheckBase) RunCheck() {
	c.Result.Status = Fail
	c.Result.Failures = append(c.Result.Failures, "not implemented")
}

// GetResult returns the value of c.Result.
func (c *CheckBase) GetResult() Result {
	return c.Result
}

// Status calculates and returns the overall result of all check results.
func (rl *ResultList) Status() CheckStatus {
	for _, r := range rl.Results {
		if r.Status == Fail {
			return Fail
		}
	}
	return Pass
}

// TableDisplay generates the tabular output for the ResultList.
func (rl *ResultList) TableDisplay(w *tabwriter.Writer) {
	var linePass, lineFail string

	if len(rl.Results) > 0 {
		fmt.Fprintf(w, "NAME\tSTATUS\tPASSES\tFAILS\n")
		for _, r := range rl.Results {
			linePass = ""
			lineFail = ""
			if len(r.Passes) > 0 {
				linePass = r.Passes[0]
			}
			if len(r.Failures) > 0 {
				lineFail = r.Failures[0]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.Status, linePass, lineFail)

			if len(r.Passes) > 1 || len(r.Failures) > 1 {
				numPasses := len(r.Passes)
				numFailures := len(r.Failures)

				// How many additional lines?
				numAddLines := numPasses
				if numFailures > numPasses {
					numAddLines = numFailures
				}

				for i := 1; i < numAddLines; i++ {
					linePass = ""
					lineFail = ""
					if numPasses > i {
						linePass = r.Passes[i]
					}
					if numFailures > i {
						lineFail = r.Failures[i]
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "", "", linePass, lineFail)
				}
			}
		}
	}

	w.Flush()
}
