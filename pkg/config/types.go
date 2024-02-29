package config

import (
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type Config struct {
	// The directory to audit.
	ProjectDir string `yaml:"project-dir"`
	// The severity level for which the program will exit with an error.
	// Default is high.
	FailSeverity Severity `yaml:"fail-severity"`
	Checks       CheckMap `yaml:"checks"`
	Remediate    bool     `yaml:"-"`
	// If requesting LagoonFact output, the base url and token for the Lagoon
	// api are required to infer environment IDs and the like.
	LagoonApiBaseUrl string `yaml:"lagoon-api-base-url"`
}

type ConfigV2 struct {
	Connections map[string]map[string]interface{} `yaml:"connections"`
	Gather      map[string]map[string]interface{} `yaml:"gather"`
	Analyse     map[string]map[string]interface{} `yaml:"analyse"`
	Output      map[string]map[string]interface{} `yaml:"output"`
}

type Severity string

const (
	LowSeverity      Severity = "low"
	NormalSeverity   Severity = "normal"
	HighSeverity     Severity = "high"
	CriticalSeverity Severity = "critical"
)

type CheckMap map[CheckType][]Check

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
	AddBreach(result.Breach)
	AddPass(msg string)
	AddWarning(msg string)
	SetPerformRemediation(flag bool)
	RunCheck()
	ShouldPerformRemediation() bool
	Remediate()
	GetResult() *result.Result
}

// CheckBase provides the basic structure for all Checks.
type CheckBase struct {
	Name  string `yaml:"name"`
	cType CheckType
	// Flag indicating if the check requires a database connection to run.
	RequiresDb bool              `yaml:"-"`
	DataMap    map[string][]byte `yaml:"-"`
	Result     result.Result     `yaml:"-"`
	// Default severity is normal.
	Severity           `yaml:"severity"`
	PerformRemediation bool `yaml:"-"`
}
