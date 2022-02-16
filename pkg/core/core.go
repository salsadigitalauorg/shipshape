package core

import (
	"fmt"
	"text/tabwriter"
)

func (c *CheckBase) Init(pd string, ct CheckType) {
	c.ProjectDir = pd
	if c.Result.CheckType == "" {
		c.Result = Result{Name: c.Name, CheckType: ct}
	}
}

func (c *CheckBase) GetName() string {
	return c.Name
}

func (c *CheckBase) FetchData() {}

func (c *CheckBase) RunCheck() {}

func (c *CheckBase) GetResult() Result {
	return c.Result
}

func (rl *ResultList) TableDisplay(w *tabwriter.Writer) {
	var linePass, lineFail string

	if len(rl.Results) > 0 {
		fmt.Fprintf(w, "NAME\tTYPE\tSTATUS\tPASSES\tFAILS\n")
		for _, r := range rl.Results {
			linePass = ""
			lineFail = ""
			if len(r.Passes) > 0 {
				linePass = r.Passes[0]
			}
			if len(r.Failures) > 0 {
				lineFail = r.Failures[0]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Name, r.CheckType, r.Status, linePass, lineFail)

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
						linePass = r.Failures[i]
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "", "", "", linePass, lineFail)
				}
			}
		}
	}

	w.Flush()
}
