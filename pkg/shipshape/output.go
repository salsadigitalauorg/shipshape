package shipshape

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"text/tabwriter"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
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
func SimpleDisplay(w *bufio.Writer) {
	if len(RunResultList.Results) == 0 {
		fmt.Fprint(w, "No result available; ensure your shipshape.yml is configured correctly.\n")
		w.Flush()
		return
	}

	printRemediations := func() {
		for _, r := range RunResultList.Results {
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

	if RunResultList.Status() == config.Pass && int(RunResultList.TotalRemediations) == 0 {
		fmt.Fprint(w, "Ship is in top shape; no breach detected!\n")
		w.Flush()
		return
	} else if RunResultList.Status() == config.Pass && int(RunResultList.TotalRemediations) > 0 {
		fmt.Fprintf(w, "Breaches were detected but were all fixed successfully!\n\n")
		printRemediations()
		w.Flush()
		return
	}

	if RunResultList.RemediationPerformed && int(RunResultList.TotalBreaches) > 0 {
		fmt.Fprint(w, "Breaches were detected but not all of them could "+
			"be fixed as they are either not supported yet or there were "+
			"errors when trying to remediate.\n\n")
		fmt.Fprint(w, "# Remediations\n\n")
		printRemediations()
		fmt.Fprint(w, "# Non-remediated breaches\n\n")
	} else if !RunResultList.RemediationPerformed {
		fmt.Fprint(w, "# Breaches were detected\n\n")
	}

	for _, r := range RunResultList.Results {
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
			Tests:     RunResultList.CheckCountByType[ct],
			Errors:    RunResultList.BreachCountByType[ct],
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
