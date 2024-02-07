package shipshape

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"text/tabwriter"

	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

// TableDisplay generates the tabular output for the ResultList.
func TableDisplay(w *tabwriter.Writer) {
	var linePass, lineFail string

	if len(RunResultList.Results) == 0 {
		fmt.Fprint(w, "No result available; ensure your shipshape.yml is configured correctly.\n")
		w.Flush()
		return
	}

	fmt.Fprintf(w, "NAME\tSTATUS\tPASSES\tFAILS\n")
	for _, r := range RunResultList.Results {
		linePass = ""
		lineFail = ""
		if len(r.Passes) > 0 {
			linePass = r.Passes[0]
		}
		if len(r.Breaches) > 0 {
			lineFail = r.Breaches[0].String()
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.Status, linePass, lineFail)

		if len(r.Passes) > 1 || len(r.Breaches) > 1 {
			numPasses := len(r.Passes)
			numFailures := len(r.Breaches)

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
					lineFail = r.Breaches[i].String()
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "", "", linePass, lineFail)
			}
		}
	}
	w.Flush()
}

// SimpleDisplay outputs only failures to the writer.
func SimpleDisplay(w *bufio.Writer) {
	if len(RunResultList.Results) == 0 {
		fmt.Fprint(w, "No result available; ensure your shipshape.yml is configured correctly.\n")
		w.Flush()
		return
	}

	printRemediations := func() {
		for _, r := range RunResultList.Results {
			_, successful, _, _ := r.RemediationsCount()
			if successful == 0 {
				continue
			}
			fmt.Fprintf(w, "  ### %s\n", r.Name)
			for _, b := range r.Breaches {
				if b.GetRemediation().Status != result.RemediationStatusSuccess {
					continue
				}
				for _, msg := range b.GetRemediation().Messages {
					fmt.Fprintf(w, "     -- %s\n", msg)
				}
			}
			fmt.Fprintln(w)
		}
	}

	if RunResultList.RemediationPerformed && RunResultList.TotalBreaches > 0 {
		switch RunResultList.RemediationStatus() {
		case result.RemediationStatusNoSupport:
			fmt.Fprint(w, "Breaches were detected but none of them could be "+
				"fixed as remediation is not supported for them yet.\n\n")
			fmt.Fprint(w, "# Non-remediated breaches\n\n")
		case result.RemediationStatusFailed:
			fmt.Fprint(w, "Breaches were detected but none of them could "+
				"be fixed as there were errors when trying to remediate.\n\n")
			fmt.Fprint(w, "# Non-remediated breaches\n\n")
		case result.RemediationStatusPartial:
			fmt.Fprint(w, "Breaches were detected but not all of them could "+
				"be fixed as they are either not supported yet or there were "+
				"errors when trying to remediate.\n\n")
			fmt.Fprint(w, "# Remediations\n\n")
			printRemediations()
			fmt.Fprint(w, "# Non-remediated breaches\n\n")
		case result.RemediationStatusSuccess:
			fmt.Fprintf(w, "Breaches were detected but were all fixed successfully!\n\n")
			printRemediations()
			w.Flush()
			return
		}
	} else if RunResultList.Status() == result.Pass {
		fmt.Fprint(w, "Ship is in top shape; no breach detected!\n")
		w.Flush()
		return
	}

	if !RunResultList.RemediationPerformed {
		fmt.Fprint(w, "# Breaches were detected\n\n")
	}

	for _, r := range RunResultList.Results {
		if len(r.Breaches) == 0 || r.RemediationStatus == result.RemediationStatusSuccess {
			continue
		}
		fmt.Fprintf(w, "  ### %s\n", r.Name)
		for _, b := range r.Breaches {
			if b.GetRemediation().Status == result.RemediationStatusSuccess {
				continue
			}
			fmt.Fprintf(w, "     -- %s\n", b)
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

// JUnit outputs the checks results in the JUnit XML format.
func JUnit(w *bufio.Writer) {
	tss := JUnitTestSuites{
		Tests:      RunResultList.TotalChecks,
		Errors:     RunResultList.TotalBreaches,
		TestSuites: []JUnitTestSuite{},
	}

	// Create a JUnitTestSuite for each CheckType.
	for ct, checks := range RunConfig.Checks {
		ts := JUnitTestSuite{
			Name:      string(ct),
			Tests:     RunResultList.CheckCountByType[string(ct)],
			Errors:    RunResultList.BreachCountByType[string(ct)],
			TestCases: []JUnitTestCase{},
		}

		// Create a JUnitTestCase for each Check.
		for _, c := range checks {
			tc := JUnitTestCase{
				Name:      c.GetName(),
				ClassName: c.GetName(),
				Errors:    []JUnitError{},
			}

			for _, b := range RunResultList.GetBreachesByCheckName(c.GetName()) {
				tc.Errors = append(tc.Errors, JUnitError{Message: b.String()})
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
