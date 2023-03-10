package shipshape

import (
	"encoding/xml"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

// ResultList is a wrapper around a list of results, providing some useful
// methods to manipulate and use it.
type ResultList struct {
	// TODO: Remove config from here, Will help remove circular
	// dependency on config.
	RemediationPerformed         bool                     `json:"remediation-performed"`
	TotalChecks                  uint32                   `json:"total-checks"`
	TotalBreaches                uint32                   `json:"total-breaches"`
	TotalRemediations            uint32                   `json:"total-remediations"`
	TotalUnsupportedRemediations uint32                   `json:"total-unsupported-remediations"`
	CheckCountByType             map[config.CheckType]int `json:"check-count-by-type"`
	BreachCountByType            map[config.CheckType]int `json:"breach-count-by-type"`
	BreachCountBySeverity        map[config.Severity]int  `json:"breach-count-by-severity"`
	RemediationCountByType       map[config.CheckType]int `json:"remediation-count-by-type"`
	Results                      []config.Result          `json:"results"`
}

type JUnitError struct {
	XMLName xml.Name `xml:"error"`
	Message string   `xml:"message,attr"`
}

type JUnitTestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	Name      string   `xml:"name,attr"`
	ClassName string   `xml:"classname,attr"`
	Errors    []JUnitError
}

type JUnitTestSuite struct {
	XMLName   xml.Name `xml:"testsuite"`
	Name      string   `xml:"name,attr"`
	Tests     int      `xml:"tests,attr"`
	Errors    int      `xml:"errors,attr"`
	TestCases []JUnitTestCase
}

// JUnit format taken from https://llg.cubic.org/docs/junit/.
type JUnitTestSuites struct {
	XMLName    xml.Name `xml:"testsuites"`
	Tests      uint32   `xml:"tests,attr"`
	Errors     uint32   `xml:"errors,attr"`
	TestSuites []JUnitTestSuite
}
