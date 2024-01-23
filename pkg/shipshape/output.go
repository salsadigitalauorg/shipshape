package shipshape

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"text/tabwriter"

	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"
	"github.com/salsadigitalauorg/shipshape/pkg/result"

	log "github.com/sirupsen/logrus"
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

	if RunResultList.Status() == result.Pass && int(RunResultList.TotalRemediations) == 0 {
		fmt.Fprint(w, "Ship is in top shape; no breach detected!\n")
		w.Flush()
		return
	} else if RunResultList.Status() == result.Pass && int(RunResultList.TotalRemediations) > 0 {
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

// LagoonProblems pushes problems to either the Lagoon API or Insights-Remote endpoint
// see https://github.com/uselagoon/insights-remote#insights-written-directly-to-insights-remote
func LagoonProblems(w *bufio.Writer) {
	problems := []lagoon.Problem{}

	if RunResultList.TotalBreaches == 0 {
		if lagoon.PushFacts {
			lagoon.InitClient()
			err := lagoon.DeleteProblems()
			if err != nil {
				log.WithError(err).Fatal("failed to delete problems")
			}
			fmt.Fprint(w, "no breach to push to Lagoon; only deleted previous problems")
			w.Flush()
			return
		}
		w.Flush()
		return
	}

	for _, r := range RunResultList.Results {
		// let's marshall the breaches, they can be attached to the problem in the data field
		_, err := json.Marshal(r.Breaches)
		if err != nil {
			log.WithError(err).Fatal("Unable to marshall breach information")
		}

		//if len(value) > lagoon.FactMaxValueLength {
		//	value = value[:lagoon.FactMaxValueLength-12] + "...TRUNCATED"
		//}
		problems = append(problems, lagoon.Problem{
			Identifier:        r.Name,
			Version:           "1",
			FixedVersion:      "",
			Source:            "shipshape",
			Service:           "",
			Data:              "{}",
			Severity:          lagoon.SeverityTranslation(config.Severity(r.Severity)),
			SeverityScore:     0,
			AssociatedPackage: "",
			Description:       "",
			Links:             "",
		})
	}

	if lagoon.PushFacts {
		lagoon.InitClient()
	}
	if lagoon.PushFactsToInsightRemote {
		// first, let's try doing this via in-cluster functionality
		bearerToken, err := lagoon.GetBearerTokenFromDisk(lagoon.DefaultLagoonInsightsTokenLocation)
		if err == nil { // we have a token, and so we can proceed via the internal service call
			err = lagoon.ProblemsToInsightsRemote(problems, lagoon.LagoonInsightsRemoteEndpoint, bearerToken)
			if err != nil {
				log.WithError(err).Fatal("Unable to write facts to Insights Remote")
			}

		} else {
			log.WithError(err).Fatal("Bearer token unable to be loaded from ", lagoon.DefaultLagoonInsightsTokenLocation)
		}
		fmt.Fprint(w, "successfully pushed facts to Lagoon Remote")
		w.Flush()
		return
	}
	w.Flush()
}

// LagoonFacts outputs breaches in a format compatible with the
// lagoon-facts-app to be consumed.
// see https://github.com/uselagoon/lagoon-facts-app#arbitrary-facts
func LagoonFacts(w *bufio.Writer) {
	facts := []lagoon.Fact{}

	if RunResultList.TotalBreaches == 0 {
		if lagoon.PushFacts {
			lagoon.InitClient()
			err := lagoon.DeleteProblems()
			if err != nil {
				log.WithError(err).Fatal("failed to delete facts")
			}
			fmt.Fprint(w, "no breach to push to Lagoon; only deleted previous facts")
			w.Flush()
			return
		}
		fmt.Fprint(w, "[]")
		w.Flush()
		return
	}

	for iR, r := range RunResultList.Results {
		for iB, b := range r.Breaches {
			value := lagoon.BreachFactValue(b)
			if len(value) > lagoon.FactMaxValueLength {
				value = value[:lagoon.FactMaxValueLength-12] + "...TRUNCATED"
			}
			facts = append(facts, lagoon.Fact{
				Name:        fmt.Sprintf("[%d] %s", iR+iB+1, lagoon.BreachFactName(b)),
				Description: result.BreachGetCheckName(b),
				Value:       value,
				Source:      lagoon.SourceName,
				Category:    string(result.BreachGetCheckType(b)),
			})
		}
	}

	if lagoon.PushFactsToInsightRemote {
		// first, let's try doing this via in-cluster functionality
		bearerToken, err := lagoon.GetBearerTokenFromDisk(lagoon.DefaultLagoonInsightsTokenLocation)
		if err == nil { // we have a token, and so we can proceed via the internal service call
			err = lagoon.FactsToInsightsRemote(facts, lagoon.LagoonInsightsRemoteEndpoint, bearerToken)
			if err != nil {
				log.WithError(err).Fatal("Unable to write facts to Insights Remote")
			}
		} else {
			log.WithError(err).Fatal("Bearer token unable to be loaded from ", lagoon.DefaultLagoonInsightsTokenLocation)
		}
		fmt.Fprint(w, "successfully pushed facts to Lagoon Remote")
		w.Flush()
		return
	}

	if lagoon.PushFacts {
		lagoon.InitClient()
		err := lagoon.ReplaceFacts(facts)
		if err != nil {
			log.WithError(err).Fatal("failed to replace facts")
		}
		fmt.Fprint(w, "successfully pushed facts to the Lagoon api")
		w.Flush()
		return
	}

	factsBytes, err := json.Marshal(facts)
	if err != nil {
		log.WithError(err).Fatalf("error occurred while converting to json")
	}
	fmt.Fprint(w, string(factsBytes))
	w.Flush()
}
