package output

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type Stdout struct {
	// Plugin-specific fields.
	// Format is the output format. One of "pretty", "table", "json".
	Format string `yaml:"format"`
}

var OutputFormats = []string{"json", "pretty", "table"}
var s = &Stdout{Format: "pretty"}

func init() {
	Outputters["stdout"] = s
}

func (p *Stdout) Output(rl *result.ResultList) ([]byte, error) {
	var buf bytes.Buffer
	switch p.Format {
	case "pretty":
		p.Pretty(rl, &buf)
	case "table":
		p.Table(rl, &buf)
	case "json":
		data, err := json.Marshal(rl)
		if err != nil {
			return nil, fmt.Errorf("unable to convert result to json: %+v", err)
		}
		fmt.Fprintln(&buf, string(data))
	case "junit":
		p.JUnit(rl, &buf)
	}
	return buf.Bytes(), nil
}

// TableDisplay generates the tabular output for the ResultList.
func (p *Stdout) Table(rl *result.ResultList, w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	var linePass, lineFail string

	if len(rl.Results) == 0 {
		fmt.Fprint(tw, "No result available; ensure your shipshape.yml is configured correctly.\n")
		tw.Flush()
		return
	}

	fmt.Fprintf(tw, "NAME\tSTATUS\tPASSES\tFAILS\n")
	for _, r := range rl.Results {
		linePass = ""
		lineFail = ""
		if len(r.Passes) > 0 {
			linePass = r.Passes[0]
		}
		if len(r.Breaches) > 0 {
			lineFail = r.Breaches[0].String()
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.Name, r.Status, linePass, lineFail)

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
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", "", "", linePass, lineFail)
			}
		}
	}
	tw.Flush()
}

// Pretty outputs only failures to the writer.
func (p *Stdout) Pretty(rl *result.ResultList, w io.Writer) {
	buf := bufio.NewWriter(w)
	if len(rl.Results) == 0 {
		fmt.Fprint(buf, "No result available; ensure your shipshape.yml is configured correctly.\n")
		buf.Flush()
		return
	}

	printRemediations := func() {
		for _, r := range rl.Results {
			_, successful, _, _ := r.RemediationsCount()
			if successful == 0 {
				continue
			}
			fmt.Fprintf(buf, "  ### %s\n", r.Name)
			for _, b := range r.Breaches {
				if b.GetRemediation().Status != breach.RemediationStatusSuccess {
					continue
				}
				for _, msg := range b.GetRemediation().Messages {
					fmt.Fprintf(buf, "     -- %s\n", msg)
				}
			}
			fmt.Fprintln(buf)
		}
	}

	if rl.RemediationPerformed && rl.TotalBreaches > 0 {
		switch rl.RemediationStatus() {
		case breach.RemediationStatusNoSupport:
			fmt.Fprint(buf, "Breaches were detected but none of them could be "+
				"fixed as remediation is not supported for them yet.\n\n")
			fmt.Fprint(buf, "# Non-remediated breaches\n\n")
		case breach.RemediationStatusFailed:
			fmt.Fprint(buf, "Breaches were detected but none of them could "+
				"be fixed as there were errors when trying to remediate.\n\n")
			fmt.Fprint(buf, "# Non-remediated breaches\n\n")
		case breach.RemediationStatusPartial:
			fmt.Fprint(buf, "Breaches were detected but not all of them could "+
				"be fixed as they are either not supported yet or there were "+
				"errors when trying to remediate.\n\n")
			fmt.Fprint(buf, "# Remediations\n\n")
			printRemediations()
			fmt.Fprint(buf, "# Non-remediated breaches\n\n")
		case breach.RemediationStatusSuccess:
			fmt.Fprintf(buf, "Breaches were detected but were all fixed successfully!\n\n")
			printRemediations()
			buf.Flush()
			return
		}
	} else if rl.Status() == result.Pass {
		fmt.Fprint(buf, "Ship is in top shape; no breach detected!\n")
		buf.Flush()
		return
	}

	if !rl.RemediationPerformed {
		fmt.Fprint(buf, "# Breaches were detected\n\n")
	}

	for _, r := range rl.Results {
		if len(r.Breaches) == 0 || r.RemediationStatus == breach.RemediationStatusSuccess {
			continue
		}
		fmt.Fprintf(buf, "  ### %s\n", r.Name)
		for _, b := range r.Breaches {
			if b.GetRemediation().Status == breach.RemediationStatusSuccess {
				continue
			}
			fmt.Fprintf(buf, "     -- %s\n", b)
		}
		fmt.Fprintln(buf)
	}
	buf.Flush()
}

// JUnit outputs the checks results in the JUnit XML format.
func (p *Stdout) JUnit(rl *result.ResultList, w io.Writer) {
	buf := bufio.NewWriter(w)
	tss := JUnitTestSuites{
		Tests:      rl.TotalChecks,
		Errors:     rl.TotalBreaches,
		TestSuites: []JUnitTestSuite{},
	}

	// Create a JUnitTestSuite for each CheckType.
	for pplugin, policies := range rl.Policies {
		ts := JUnitTestSuite{
			Name:      pplugin,
			Tests:     rl.CheckCountByType[pplugin],
			Errors:    rl.BreachCountByType[pplugin],
			TestCases: []JUnitTestCase{},
		}

		// Create a JUnitTestCase for each Check.
		for _, plc := range policies {
			tc := JUnitTestCase{
				Name:      plc,
				ClassName: plc,
				Errors:    []JUnitError{},
			}

			for _, b := range rl.GetBreachesByCheckName(plc) {
				tc.Errors = append(tc.Errors, JUnitError{Message: b.String()})
			}
			ts.TestCases = append(ts.TestCases, tc)
		}
		tss.TestSuites = append(tss.TestSuites, ts)
	}

	xmlBytes, err := xml.MarshalIndent(tss, "", "    ")
	if err != nil {
		fmt.Fprintf(buf, "error occurred while converting to XML: %s\n", err.Error())
		buf.Flush()
		return
	}
	fmt.Fprint(buf, xml.Header)
	fmt.Fprint(buf, string(xmlBytes))
	fmt.Fprintln(buf)
	buf.Flush()
}
