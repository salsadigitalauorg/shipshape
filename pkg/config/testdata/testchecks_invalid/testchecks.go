package testchecks_invalid

import (
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

const TestCheckInvalid config.CheckType = "test-check (2)"

type TestCheckInvalidCheck struct {
	config.CheckBase `yaml:",inline"`
	Zoom             string `yaml:"zap"`
}

func RegisterChecks() {
	config.ChecksRegistry[TestCheckInvalid] = func() config.Check { return &TestCheckInvalidCheck{} }
}
