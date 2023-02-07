package testchecks

import "github.com/salsadigitalauorg/shipshape/pkg/shipshape"

const TestCheck1 shipshape.CheckType = "test-check-1"
const TestCheck2 shipshape.CheckType = "test-check-2"

type TestCheck1Check struct {
	shipshape.CheckBase `yaml:",inline"`
	Foo                 string `yaml:"foo"`
}

type TestCheck2Check struct {
	shipshape.CheckBase `yaml:",inline"`
	Bar                 string `yaml:"bar"`
}

func RegisterChecks() {
	shipshape.ChecksRegistry[TestCheck1] = func() shipshape.Check { return &TestCheck1Check{} }
	shipshape.ChecksRegistry[TestCheck2] = func() shipshape.Check { return &TestCheck2Check{} }
}
