package testchecks

import (
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

const TestCheck1 config.CheckType = "test-check-1"
const TestCheck2 config.CheckType = "test-check-2"
const TestCheck3 config.CheckType = "test-check-3"

type TestCheck1Check struct {
	config.CheckBase `yaml:",inline"`
	Foo              string `yaml:"foo"`
}

func (*TestCheck1Check) RequiresData() bool { return false }

// Merge implementation for test-check-1 check.
func (c *TestCheck1Check) Merge(mergeCheck config.Check) error {
	testCheck1MergeCheck := mergeCheck.(*TestCheck1Check)
	if err := c.CheckBase.Merge(&testCheck1MergeCheck.CheckBase); err != nil {
		return err
	}

	utils.MergeString(&c.Foo, testCheck1MergeCheck.Foo)
	return nil
}

type TestCheck2Check struct {
	config.CheckBase `yaml:",inline"`
	Bar              string `yaml:"bar"`
}

type TestCheck3Check struct {
	config.CheckBase `yaml:",inline"`
	Bar              string `yaml:"bar"`
}

func RegisterChecks() {
	config.ChecksRegistry[TestCheck1] = func() config.Check { return &TestCheck1Check{} }
	config.ChecksRegistry[TestCheck2] = func() config.Check { return &TestCheck2Check{} }
	config.ChecksRegistry[TestCheck3] = func() config.Check { return &TestCheck3Check{} }
}
