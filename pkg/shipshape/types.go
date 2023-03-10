package shipshape

import (
	"encoding/xml"
)

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
