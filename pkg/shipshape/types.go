package shipshape

import (
	"encoding/xml"

	"gopkg.in/yaml.v3"
)

type CheckType string

// Check should be implemented by any new check that has to be run in an audit.
// A number of the functions have a basic implementation in CheckBase; they can
// be used as-is, or overridden as required.
type Check interface {
	Init(pd string, ct CheckType)
	GetName() string
	RequiresData() bool
	RequiresDatabase() bool
	HasData(failCheck bool) bool
	FetchData()
	UnmarshalDataMap()
	AddFail(msg string)
	AddPass(msg string)
	RunCheck()
	GetResult() *Result
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
	Name string `yaml:"name"`
	// Flag indicating if the check requires a database connection to run.
	RequiresDb bool
	DataMap    map[string][]byte
	Result     Result
	// Default severity is normal.
	Severity Severity `yaml:"severity"`
}

// Result provides the structure for a Check's outcome.
type Result struct {
	Name      string      `json:"name"`
	Severity  Severity    `json:"severity"`
	CheckType CheckType   `json:"check-type"`
	Status    CheckStatus `json:"status"`
	Passes    []string    `json:"passes"`
	Failures  []string    `json:"failures"`
}

// ResultList is a wrapper around a list of results, providing some useful
// methods to manipulate and use it.
type ResultList struct {
	config                *Config
	TotalChecks           int               `json:"total-checks"`
	TotalBreaches         int               `json:"total-breaches"`
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

// KeyValue represents a check to be made against Yaml data.
// It can be a simple Key=Value check, or it can be a Key in DisallowedList
// check, in which case IsList needs to be true, and a Disallowed list of values
// is required.
type KeyValue struct {
	Key        string   `yaml:"key"`
	Value      string   `yaml:"value"`
	IsList     bool     `yaml:"is-list"`
	Disallowed []string `yaml:"disallowed"`
}

// KeyValueResult represents the different outcomes of the KeyValue check.
type KeyValueResult int8

const (
	KeyValueError           KeyValueResult = -2
	KeyValueNotFound        KeyValueResult = -1
	KeyValueNotEqual        KeyValueResult = 0
	KeyValueEqual           KeyValueResult = 1
	KeyValueDisallowedFound KeyValueResult = 2
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
	Path              string `yaml:"path"`
	DisallowedPattern string `yaml:"disallowed-pattern"`
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
	IgnoreMissing  bool     `yaml:"ignore-missing"`  // Allows non-existent files to not be counted as a Fail
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
	Tests      int      `xml:"tests,attr"`
	Errors     int      `xml:"errors,attr"`
	TestSuites []JUnitTestSuite
}
