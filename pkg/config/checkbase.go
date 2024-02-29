package config

import (
	"fmt"

	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

// Init acts as the constructor of a check and sets some initial values.
func (c *CheckBase) Init(ct CheckType) {
	// Default severity is normal.
	if c.Severity == "" {
		c.Severity = NormalSeverity
	}
	if c.Result.CheckType == "" {
		c.Result = result.Result{Name: c.Name, CheckType: string(ct)}
	}
	if c.Result.Severity == "" {
		c.Result.Severity = string(c.Severity)
	}

	if c.cType == "" {
		c.cType = ct
	}
}

// GetName returns the name of a check.
func (c *CheckBase) GetName() string { return c.Name }

// GetType returns the type of a check.
func (c *CheckBase) GetType() CheckType { return c.cType }

// GetSeverity returns the severity of a check.
func (c *CheckBase) GetSeverity() Severity { return c.Severity }

// Merge merges values from another check into this one.
func (c *CheckBase) Merge(mergeCheck Check) error {
	// Empty name means the merge will be done for all checks of the same type.
	if mergeCheck.GetName() != "" && c.Name != mergeCheck.GetName() {
		return fmt.Errorf("can only merge checks with the same name")
	}
	if mergeCheck.GetSeverity() != "" {
		c.Severity = mergeCheck.GetSeverity()
	}
	return nil
}

// RequiresData indicates whether the check requires a DataMap to run against.
// It is designed as opt-out, so remember to set it to false if you are creating
// a check that does not require the DataMap.
func (c *CheckBase) RequiresData() bool { return true }

// RequiresDb indicates whether the check requires a database to run against.
func (c *CheckBase) RequiresDatabase() bool { return c.RequiresDb }

// FetchData contains the logic for fetching the data over which the check is
// going to run.
// This is where c.DataMap should be populated.
func (c *CheckBase) FetchData() {}

// HasData determines whether the dataMap has been populated or not.
// The Check can optionally be marked as failed if the dataMap is not populated.
func (c *CheckBase) HasData(failCheck bool) bool {
	if c.DataMap == nil {
		if failCheck {
			c.AddBreach(&result.ValueBreach{Value: "no data available"})
		}
		return false
	}
	return true
}

// UnmarshalDataMap attempts to parse the DataMap into a structure that
// can be used to execute the check. Any failure here should fail the check.
func (c *CheckBase) UnmarshalDataMap() {}

// AddBreach appends a Breach to the Result.
func (c *CheckBase) AddBreach(b result.Breach) {
	b.SetCommonValues(string(c.cType), c.Name, string(c.Severity))
	c.Result.Breaches = append(
		c.Result.Breaches,
		b,
	)
}

// AddPass appends a Pass to the Result.
func (c *CheckBase) AddPass(msg string) {
	c.Result.Passes = append(
		c.Result.Passes,
		msg,
	)
}

// AddWarning appends a Warning message to the result.
func (c *CheckBase) AddWarning(msg string) {
	c.Result.Warnings = append(c.Result.Warnings, msg)
}

// SetPerformRemediation sets the flag for whether to remediate or not.
func (c *CheckBase) SetPerformRemediation(flag bool) {
	c.PerformRemediation = flag
}

// RunCheck contains the core logic for running the check,
// generating the result and remediating breaches.
// This is where c.Result should be populated.
func (c *CheckBase) RunCheck() {
	c.AddBreach(&result.ValueBreach{Value: "not implemented"})
}

// ShouldPerformRemediation returns whether to remediate or not.
func (c *CheckBase) ShouldPerformRemediation() bool {
	return c.PerformRemediation
}

// Remediate should implement the logic to fix the breach(es).
// Any type or custom struct can be used for the breach; it just needs to be
// cast to the required type before being used.
func (c *CheckBase) Remediate() {
	for _, b := range c.Result.Breaches {
		b.SetRemediation(result.RemediationStatusNoSupport, "")
	}
}

// GetResult returns a ref of the result.
func (c *CheckBase) GetResult() *result.Result {
	return &c.Result
}
