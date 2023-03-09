package config

type Config struct {
	// The directory to audit.
	ProjectDir string `yaml:"project-dir"`
	// The severity level for which the program will exit with an error.
	// Default is high.
	FailSeverity Severity `yaml:"fail-severity"`
	Checks       CheckMap `yaml:"checks"`
	Remediate    bool     `yaml:"-"`
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
	AddFail(msg string)
	AddPass(msg string)
	AddWarning(msg string)
	SetPerformRemediation(flag bool)
	AddRemediation(msg string)
	RunCheck()
	GetResult() *Result
	Remediate(breachIfc interface{}) error
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

type CheckStatus string

const (
	Pass CheckStatus = "Pass"
	Fail CheckStatus = "Fail"
)

// CheckBase provides the basic structure for all Checks.
type CheckBase struct {
	Name  string `yaml:"name"`
	cType CheckType
	// Flag indicating if the check requires a database connection to run.
	RequiresDb bool              `yaml:"-"`
	DataMap    map[string][]byte `yaml:"-"`
	Result     Result            `yaml:"-"`
	// Default severity is normal.
	Severity           Severity `yaml:"severity"`
	PerformRemediation bool     `yaml:"-"`
}
