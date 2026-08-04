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
	// TestError, when set, is recorded as a Collect() error instead of
	// setting data - used to exercise error-propagation in the manager.
	TestError error
	// TestSupportedInputFormatsLevel, when non-empty, overrides
	// SupportedInputFormats() so a TestFacter can stand in for a fact that
	// actually declares an input requirement (e.g. to exercise
	// ValidateInput()/CollectFact() gating). Left unset, the zero value
	// preserves the original SupportNone/no-formats behaviour every existing
	// caller relies on.
	TestSupportedInputFormatsLevel plugin.SupportLevel
	TestSupportedInputFormatsList  []data.DataFormat
	// TestCollected records whether Collect() has run, so tests can assert
	// directly on invocation rather than inferring it from GetData()/
	// GetFormat() (which are also both zero-valued before Collect() runs,
	// making "not collected" and "collected but produced a zero value"
	// indistinguishable without this).
	TestCollected bool
}

func init() {
	fact.Manager().RegisterFactory("testdata:testfacter", func(n string) fact.Facter {
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

// SupportedInputFormats defaults to plugin.SupportNone/no formats (the
// original TestFacter behaviour), unless TestSupportedInputFormatsLevel has
// been set to opt a specific test into input validation.
func (p *TestFacter) SupportedInputFormats() (plugin.SupportLevel, []data.DataFormat) {
	if p.TestSupportedInputFormatsLevel == "" {
		return plugin.SupportNone, []data.DataFormat{}
	}
	return p.TestSupportedInputFormatsLevel, p.TestSupportedInputFormatsList
}

func (p *TestFacter) Collect() {
	p.TestCollected = true
	// Format is set regardless of TestError, mirroring real fact plugins
	// (e.g. file:read sets Format at construction time via NewRead, entirely
	// independent of whether Collect() later errors). This matters for
	// tests that assert on ValidateInput()'s inPlug.GetFormat() == "" check,
	// which would otherwise misreport an errored-but-well-formed input as
	// unformatted.
	p.Format = p.TestInputDataFormat
	if p.TestError != nil {
		p.AddErrors(p.TestError)
		return
	}
	p.SetData(p.TestInputData)
}
