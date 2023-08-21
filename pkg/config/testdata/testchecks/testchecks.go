package testchecks

import (
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

const TestCheck1 config.CheckType = "test-check-1"
const TestCheck2 config.CheckType = "test-check-2"

type TestCheck1Check struct {
	config.CheckBase `yaml:",inline"`
	Foo              string `yaml:"foo"`
}

func (*TestCheck1Check) RequiresData() bool { return false }

type TestCheck2Check struct {
	config.CheckBase `yaml:",inline"`
	Bar              string `yaml:"bar"`
}

func RegisterChecks() {
	config.ChecksRegistry[TestCheck1] = func() config.Check { return &TestCheck1Check{} }
	config.ChecksRegistry[TestCheck2] = func() config.Check { return &TestCheck2Check{} }
}
