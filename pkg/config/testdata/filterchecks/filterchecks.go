package filterchecks

import (
	"github.com/salsadigitalauorg/shipshape/pkg/config"
)

const FilterCheck1 config.CheckType = "filter-check-1"
const FilterCheck2 config.CheckType = "filter-check-2"

type FilterCheck1Check struct {
	config.CheckBase `yaml:",inline"`
	Foo              string `yaml:"foo"`
}

func (c *FilterCheck1Check) Init(ct config.CheckType) {
	c.CheckBase.Init(ct)
	c.RequiresDb = true
}

type FilterCheck2Check struct {
	config.CheckBase `yaml:",inline"`
	Bar              string `yaml:"bar"`
}

func RegisterChecks() {
	config.ChecksRegistry[FilterCheck1] = func() config.Check { return &FilterCheck1Check{} }
	config.ChecksRegistry[FilterCheck2] = func() config.Check { return &FilterCheck2Check{} }
}
