package shipshape

import (
	"encoding/xml"

	"gopkg.in/yaml.v3"
)

var ProjectDir string

type CheckType string

// Check should be implemented by any new check that has to be run in an audit.
// A number of the functions have a basic implementation in CheckBase; they can
// be used as-is, or overridden as required.
type Check interface {
	Init(ct CheckType)
	GetName() string
	GetType() CheckType
	GetSeverity() Severity
	Merge(Check) error
	RequiresData() bool
	RequiresDatabase() bool
	HasData(failCheck bool) bool
	FetchData()
	UnmarshalDataMap()
	AddFail(msg string)
	AddPass(msg string)
	AddWarning(msg string)
	AddRemediation(msg string)
	RunCheck(remediate bool)
	GetResult() *Result
	Remediate(breachIfc interface{}) error
}

type CheckMap map[CheckType][]Check

type Severity string

const (
	LowSeverity      Severity = "low"
	NormalSeverity   Severity = "normal"
	HighSeverity     Severity = "high"
	CriticalSeverity Severity = "critical"
)

type Config struct {
	// The directory to audit.
	ProjectDir string `yaml:"project-dir"`
	// The severity level for which the program will exit with an error.
	// Default is high.
	FailSeverity Severity `yaml:"fail-severity"`
	Checks       CheckMap `yaml:"checks"`
}

// CheckBase provides the basic structure for all Checks.
type CheckBase struct {
	Name  string `yaml:"name"`
	cType CheckType
	// Flag indicating if the check requires a database connection to run.
	RequiresDb bool              `yaml:"-"`
	DataMap    map[string][]byte `yaml:"-"`
	Result     Result            `yaml:"-"`
	// Default severity is normal.
	Severity Severity `yaml:"severity"`
}

// Result provides the structure for a Check's outcome.
type Result struct {
	Name         string `json:"name"`
	Severity     `json:"severity"`
	CheckType    `json:"check-type"`
	Status       CheckStatus `json:"status"`
	Passes       []string    `json:"passes"`
	Failures     []string    `json:"failures"`
	Warnings     []string    `json:"warnings"`
	Remediations []string    `json:"remediations"`
}

// ResultList is a wrapper around a list of results, providing some useful
// methods to manipulate and use it.
type ResultList struct {
	config                *Config
	TotalChecks           uint32            `json:"total-checks"`
	TotalBreaches         uint32            `json:"total-breaches"`
	CheckCountByType      map[CheckType]int `json:"check-count-by-type"`
	BreachCountByType     map[CheckType]int `json:"breach-count-by-type"`
	BreachCountBySeverity map[Severity]int  `json:"breach-count-by-severity"`
	Results               []Result          `json:"results"`
}

var OutputFormats = []string{"json", "junit", "simple", "table"}

type CheckStatus string

const (
	Pass CheckStatus = "Pass"
	Fail CheckStatus = "Fail"
)

const (
	File     CheckType = "file"     // Represents a FileCheck.
	Yaml     CheckType = "yaml"     // Represents a YamlCheck.
	YamlLint CheckType = "yamllint" // Represents a YamlLintCheck.
	Crawler  CheckType = "crawler"  // Represents a CrawlerCheck
)

// FileCheck is a simple File absence check which can be for a single
// file or a pattern.
type FileCheck struct {
	CheckBase         `yaml:",inline"`
	Path              string   `yaml:"path"`
	DisallowedPattern string   `yaml:"disallowed-pattern"`
	ExcludePattern    string   `yaml:"exclude-pattern"`
	SkipDir           []string `yaml:"skip-dir"`
}

// CrawlerCheck is a lightweight crawler that can be used to determine
// health of the project.
type CrawlerCheck struct {
	CheckBase    `yaml:",inline"`
	Domain       string   `yaml:"domain"`
	ExtraDomains []string `yaml:"extra_domains"`
	IncludeURLs  []string `yaml:"include_urls"`
	Limit        int      `yaml:"limit"`
}

// YamlBase represents the structure for a Yaml-based check.
type YamlBase struct {
	CheckBase `yaml:",inline"`
	Values    []KeyValue `yaml:"values"`
	Node      yaml.Node
	NodeMap   map[string]yaml.Node
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

var ChecksRegistry = map[CheckType]func() Check{
	File:     func() Check { return &FileCheck{} },
	Yaml:     func() Check { return &YamlCheck{} },
	YamlLint: func() Check { return &YamlLintCheck{} },
	Crawler:  func() Check { return &CrawlerCheck{} },
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
