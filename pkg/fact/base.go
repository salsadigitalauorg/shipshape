package fact

import (
	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

// BaseFact provides common fields and functionality for fact plugins.
type BaseFact struct {
	plugin.BasePlugin
	Format               data.DataFormat `yaml:"format"`
	ConnectionName       string          `yaml:"connection"`
	InputName            string          `yaml:"input"`
	AdditionalInputNames []string        `yaml:"additional-inputs"`

	connection       connection.Connectioner
	input            Facter
	additionalInputs []Facter
	data             interface{}
}

// Common getter methods
func (p *BaseFact) GetName() string {
	return p.GetId()
}

func (p *BaseFact) GetFormat() data.DataFormat {
	return p.Format
}

func (p *BaseFact) GetConnectionName() string {
	return p.ConnectionName
}

func (p *BaseFact) GetConnection() connection.Connectioner {
	return p.connection
}

func (p *BaseFact) GetInputName() string {
	return p.InputName
}

func (p *BaseFact) GetInput() Facter {
	return p.input
}

func (p *BaseFact) GetAdditionalInputNames() []string {
	return p.AdditionalInputNames
}

func (p *BaseFact) GetAdditionalInputs() []Facter {
	return p.additionalInputs
}

func (p *BaseFact) GetErrors() []error {
	if p.input != nil {
		p.AddErrors(p.input.GetErrors()...)
		return p.BasePlugin.GetErrors()
	}
	return p.BasePlugin.GetErrors()
}

func (p *BaseFact) GetData() interface{} {
	return p.data
}

func (p *BaseFact) SetInputName(name string) {
	p.InputName = name
}

func (p *BaseFact) SetData(data interface{}) {
	p.data = data
}

// Default implementations for support methods
func (p *BaseFact) SupportedConnections() (plugin.SupportLevel, []string) {
	return plugin.SupportNone, []string{}
}

func (p *BaseFact) SupportedInputs() (plugin.SupportLevel, []string) {
	return plugin.SupportNone, []string{}
}

func (p *BaseFact) ValidateConnection() error {
	connectionSupport, supportedConnections := p.SupportedConnections()

	log.WithFields(log.Fields{
		"fact":                  p.GetName(),
		"connection-support":    connectionSupport,
		"supported-connections": supportedConnections,
	}).Debug("validating connection")

	if (connectionSupport == plugin.SupportOptional ||
		connectionSupport == plugin.SupportNone) &&
		len(supportedConnections) == 0 && p.GetConnectionName() == "" {
		return nil
	}

	if connectionSupport == plugin.SupportRequired && p.GetConnectionName() == "" {
		return &plugin.ErrSupportRequired{
			Plugin: p.GetName(), SupportType: "connection"}
	}

	connPlug := connection.GetInstance(p.GetConnectionName())
	if connPlug == nil {
		return &plugin.ErrSupportNotFound{
			Plugin:        p.GetName(),
			SupportType:   "connection",
			SupportPlugin: p.GetConnectionName()}
	}

	for _, s := range supportedConnections {
		if connPlug.PluginName() == s {
			p.connection = connPlug
			return nil
		}
	}
	return &plugin.ErrSupportNone{
		Plugin:        p.GetName(),
		SupportType:   "connection",
		SupportPlugin: connPlug.PluginName()}
}

func (p *BaseFact) ValidateInput() error {
	inputSupport, supportedInputs := p.SupportedInputs()
	log.WithFields(log.Fields{
		"fact":             p.GetName(),
		"input-support":    inputSupport,
		"supported-inputs": supportedInputs,
	}).Debug("validating input")

	if (inputSupport == plugin.SupportOptional ||
		inputSupport == plugin.SupportNone) &&
		len(supportedInputs) == 0 && p.GetInputName() == "" {
		return nil
	}

	if inputSupport == plugin.SupportRequired && p.GetInputName() == "" {
		return &plugin.ErrSupportRequired{Plugin: p.GetName(), SupportType: "input"}
	}

	if p.GetInputName() != "" {
		inPlug := GetInstance(p.GetInputName())
		if inPlug == nil {
			return &plugin.ErrSupportNotFound{
				Plugin:        p.GetName(),
				SupportType:   "input",
				SupportPlugin: p.GetInputName()}
		}

		log.WithFields(log.Fields{
			"fact":                p.GetName(),
			"input-plugin":        inPlug.GetName(),
			"input-plugin-format": inPlug.GetFormat(),
		}).Debug("found input plugin")

		if inPlug.GetFormat() == "" {
			return &plugin.ErrSupportRequired{
				Plugin: inPlug.GetName(), SupportType: "input data format"}
		}

		for _, s := range supportedInputs {
			if inPlug.GetName() == s {
				p.input = inPlug
				return nil
			}
		}

		return &plugin.ErrSupportNone{
			SupportType:   "input",
			SupportPlugin: inPlug.GetName(),
			Plugin:        p.GetName(),
		}
	}

	return &plugin.ErrSupportNotFound{
		SupportType:   "input",
		Plugin:        p.GetName(),
		SupportPlugin: p.GetInputName()}
}

func (p *BaseFact) LoadAdditionalInputs() []error {
	log.WithFields(log.Fields{"fact": p.GetName()}).
		Debug("loading additional inputs")

	if len(p.GetAdditionalInputNames()) == 0 {
		return nil
	}

	plugins := []Facter{}
	errs := []error{}
	for _, n := range p.GetAdditionalInputNames() {
		inPlug := GetInstance(n)
		if inPlug == nil {
			errs = append(errs, &plugin.ErrSupportNotFound{
				Plugin:        p.GetName(),
				SupportType:   "additional input",
				SupportPlugin: n,
			})
			continue
		}

		if inPlug.GetFormat() == "" {
			errs = append(errs, &plugin.ErrSupportRequired{
				Plugin:      inPlug.GetName(),
				SupportType: "additional input data format"})
			continue
		}

		plugins = append(plugins, inPlug)
	}

	if len(errs) > 0 {
		return errs
	}

	p.additionalInputs = plugins
	return nil
}

// Collect is the main method for collecting data from the fact.
func (p *BaseFact) Collect() {}
