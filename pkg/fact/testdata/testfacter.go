package testdata

import (
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

type TestFacter struct {
	fact.BaseFact

	// Plugin fields.
	TestInputDataFormat data.DataFormat
	TestInputData       any
}

func init() {
	fact.GetManager().Register("testdata:testfacter", func(n string) fact.Facter {
		return New(n, data.FormatNil, nil)
	})
}

func New(id string, dataFormat data.DataFormat, data any) *TestFacter {
	return &TestFacter{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
		TestInputDataFormat: dataFormat,
		TestInputData:       data,
	}
}

func (p *TestFacter) GetName() string {
	return "testdata:testfacter"
}

func (p *TestFacter) Collect() {
	p.Format = p.TestInputDataFormat
	p.SetData(p.TestInputData)
}
