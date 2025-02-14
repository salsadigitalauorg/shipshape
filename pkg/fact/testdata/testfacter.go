package testdata

import (
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

type TestFacter struct {
	fact.BaseFact
	// Common fields.
	Name                 string          `yaml:"name"`
	Format               data.DataFormat `yaml:"format"`
	ConnectionName       string          `yaml:"connection"`
	InputName            string          `yaml:"input"`
	AdditionalInputNames []string        `yaml:"additional-inputs"`
	errors               []error
	data                 interface{}

	// Plugin fields.
	TestInputDataFormat data.DataFormat
	TestInputData       any
}

func init() {
	fact.Registry["file:read"] = func(n string) fact.Facter { return &TestFacter{Name: n} }
}

func (p *TestFacter) PluginName() string {
	return "file:read"
}

func (p *TestFacter) SupportedConnections() (plugin.SupportLevel, []string) {
	return plugin.SupportNone, []string{}
}

func (p *TestFacter) SupportedInputs() (plugin.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *TestFacter) Collect() {
	p.Format = p.TestInputDataFormat
	p.data = p.TestInputData
}

// Generated methods.
func (p *TestFacter) GetName() string {
	return p.Name
}

func (p *TestFacter) GetData() interface{} {
	return p.data
}

func (p *TestFacter) GetFormat() data.DataFormat {
	return p.Format
}

func (p *TestFacter) GetConnectionName() string {
	return p.ConnectionName
}

func (p *TestFacter) GetInputName() string {
	return p.InputName
}

func (p *TestFacter) GetAdditionalInputNames() []string {
	return p.AdditionalInputNames
}

func (p *TestFacter) GetErrors() []error {
	return p.errors
}

func (p *TestFacter) ValidateConnection() error {
	return &plugin.ErrSupportNone{SupportType: "connection"}
}

func (p *TestFacter) ValidateInput() error {
	return &plugin.ErrSupportNone{SupportType: "input"}
}

func (p *TestFacter) LoadAdditionalInputs() []error {
	return []error{}
}

func (p *TestFacter) AddError(err error) {
	p.errors = append(p.errors, err)
}
