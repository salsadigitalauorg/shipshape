package shipshape

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"text/tabwriter"
)

// Use locks to make map mutations concurrency-safe.
var lock = sync.RWMutex{}

func NewResultList(remediate bool) ResultList {
	return ResultList{
		RemediationPerformed:   remediate,
		Results:                []Result{},
		CheckCountByType:       map[CheckType]int{},
		BreachCountByType:      map[CheckType]int{},
		BreachCountBySeverity:  map[Severity]int{},
		RemediationCountByType: map[CheckType]int{},
	}
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

// IncrChecks increments the total checks count & checks count by type.
func (rl *ResultList) IncrChecks(ct CheckType, incr int) {
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
func (rl *ResultList) GetBreachesBySeverity(s Severity) []string {
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

// TableDisplay generates the tabular output for the ResultList.
func (rl *ResultList) TableDisplay(w *tabwriter.Writer) {
	var linePass, lineFail string

	if len(rl.Results) == 0 {
		fmt.Fprint(w, "No result available; ensure your shipshape.yml is configured correctly.\n")
		w.Flush()
		return
	}

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
	w.Flush()
}

// SimpleDisplay outputs only failures to the writer.
// TODO: Move display funcs out of ResultList to non-struct Shipshape funcs
// instead - will help remove circular dependencies.
func (rl *ResultList) SimpleDisplay(w *bufio.Writer) {
	if len(rl.Results) == 0 {
		fmt.Fprint(w, "No result available; ensure your shipshape.yml is configured correctly.\n")
		w.Flush()
		return
	}

	printRemediations := func() {
		for _, r := range rl.Results {
			if len(r.Remediations) == 0 {
				continue
			}
			fmt.Fprintf(w, "  ### %s\n", r.Name)
			for _, f := range r.Remediations {
				fmt.Fprintf(w, "     -- %s\n", f)
			}
			fmt.Fprintln(w)
		}
	}

	if rl.Status() == Pass && int(rl.TotalRemediations) == 0 {
		fmt.Fprint(w, "Ship is in top shape; no breach detected!\n")
		w.Flush()
		return
	} else if rl.Status() == Pass && int(rl.TotalRemediations) > 0 {
		fmt.Fprintf(w, "Breaches were detected but were all fixed successfully!\n\n")
		printRemediations()
		w.Flush()
		return
	}

	if rl.RemediationPerformed && int(rl.TotalBreaches) > 0 {
		fmt.Fprint(w, "Breaches were detected but not all of them could "+
			"be fixed as they are either not supported yet or there were "+
			"errors when trying to remediate.\n\n")
		fmt.Fprint(w, "# Remediations\n\n")
		printRemediations()
		fmt.Fprint(w, "# Non-remediated breaches\n\n")
	} else if !rl.RemediationPerformed {
		fmt.Fprint(w, "# Breaches were detected\n\n")
	}

	for _, r := range rl.Results {
		if len(r.Failures) == 0 {
			continue
		}
		fmt.Fprintf(w, "  ### %s\n", r.Name)
		for _, f := range r.Failures {
			fmt.Fprintf(w, "     -- %s\n", f)
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

// JUnit outputs the checks results in the JUnit XML format.
// TODO: Move display funcs out of ResultList to non-struct Shipshape funcs
// instead - will help remove circular dependencies.
func (rl *ResultList) JUnit(w *bufio.Writer) {
	tss := JUnitTestSuites{
		Tests:      rl.TotalChecks,
		Errors:     rl.TotalBreaches,
		TestSuites: []JUnitTestSuite{},
	}

	// Create a JUnitTestSuite for each CheckType.
	for ct, checks := range RunConfig.Checks {
		ts := JUnitTestSuite{
			Name:      string(ct),
			Tests:     rl.CheckCountByType[ct],
			Errors:    rl.BreachCountByType[ct],
			TestCases: []JUnitTestCase{},
		}

		// Create a JUnitTestCase for each Check.
		for _, c := range checks {
			tc := JUnitTestCase{
				Name:      c.GetName(),
				ClassName: c.GetName(),
				Errors:    []JUnitError{},
			}

			for _, b := range rl.GetBreachesByCheckName(c.GetName()) {
				tc.Errors = append(tc.Errors, JUnitError{Message: b})
			}
			ts.TestCases = append(ts.TestCases, tc)
		}
		tss.TestSuites = append(tss.TestSuites, ts)
	}

	xmlBytes, err := xml.MarshalIndent(tss, "", "    ")
	if err != nil {
		fmt.Fprintf(w, "error occurred while converting to XML: %s\n", err.Error())
		w.Flush()
		return
	}
	fmt.Fprint(w, xml.Header)
	fmt.Fprint(w, string(xmlBytes))
	fmt.Fprintln(w)
	w.Flush()
}
