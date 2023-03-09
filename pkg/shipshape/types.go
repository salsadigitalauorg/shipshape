package shipshape

import (
	"encoding/xml"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"gopkg.in/yaml.v3"
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

const (
	File     config.CheckType = "file"     // Represents a FileCheck.
	Yaml     config.CheckType = "yaml"     // Represents a YamlCheck.
	YamlLint config.CheckType = "yamllint" // Represents a YamlLintCheck.
	Crawler  config.CheckType = "crawler"  // Represents a CrawlerCheck
)

// FileCheck is a simple File absence check which can be for a single
// file or a pattern.
type FileCheck struct {
	config.CheckBase  `yaml:",inline"`
	Path              string   `yaml:"path"`
	DisallowedPattern string   `yaml:"disallowed-pattern"`
	ExcludePattern    string   `yaml:"exclude-pattern"`
	SkipDir           []string `yaml:"skip-dir"`
}

// CrawlerCheck is a lightweight crawler that can be used to determine
// health of the project.
type CrawlerCheck struct {
	config.CheckBase `yaml:",inline"`
	Domain           string   `yaml:"domain"`
	ExtraDomains     []string `yaml:"extra_domains"`
	IncludeURLs      []string `yaml:"include_urls"`
	Limit            int      `yaml:"limit"`
}

// YamlBase represents the structure for a Yaml-based check.
type YamlBase struct {
	config.CheckBase `yaml:",inline"`
	Values           []KeyValue `yaml:"values"`
	Node             yaml.Node
	NodeMap          map[string]yaml.Node
}

// YamlCheck represents a Yaml file-based check, which can be for a single file
// or across a number of files defined by a regex pattern.
type YamlCheck struct {
	YamlBase       `yaml:",inline"`
	Path           string   `yaml:"path"`            // The directory in which to lookup files.
	File           string   `yaml:"file"`            // Single file name.
	Files          []string `yaml:"files"`           // A list of files to lint.
	Pattern        string   `yaml:"pattern"`         // Pattern-based files.
	ExcludePattern string   `yaml:"exclude-pattern"` // Pattern-based excluded files.

	// IgnoreMissing allows non-existent files to not be counted as a Fail.
	// Using a pointer here so we can differentiate between
	// false (default value) and an empty value.
	IgnoreMissing *bool `yaml:"ignore-missing"`
}

// YamlLintCheck represents a Yaml lint file-based check for a number of files.
type YamlLintCheck struct {
	YamlCheck `yaml:",inline"`
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
