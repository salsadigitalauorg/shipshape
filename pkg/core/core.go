package core

import (
	"fmt"
	"text/tabwriter"
)

func (c *CheckBase) Init(pd string, ct CheckType) {
	c.ProjectDir = pd
	if c.Result.CheckType == "" {
		c.Result = Result{CheckType: ct}
	}
}

func (c *CheckBase) GetName() string {
	return c.Name
}

func (c *CheckBase) FetchData() error {
	return nil
}

func (c *CheckBase) RunCheck() error {
	return nil
}

func (c *CheckBase) GetResult() Result {
	return c.Result
}

func (rl *ResultList) TableDisplay(w *tabwriter.Writer) {
	var linePass, lineFail string

	fmt.Fprintf(w, "Name\tType\tStatus\tPasses\tFails\tError\n")
	for n, r := range rl.Results {
		linePass = ""
		lineFail = ""
		if len(r.Passes) > 0 {
			linePass = r.Passes[0]
		}
		if len(r.Failures) > 0 {
			lineFail = r.Failures[0]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", n, r.CheckType, r.Status, linePass, lineFail, r.Error)

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
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "", "", "", linePass, lineFail, "")
			}
		}
	}
	w.Flush()
}
