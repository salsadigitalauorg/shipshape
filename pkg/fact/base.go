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

func (p *BaseFact) SetConnection(conn connection.Connectioner) {
	p.connection = conn
}

func (p *BaseFact) SetInputName(name string) {
	p.InputName = name
}

func (p *BaseFact) SetInput(inP Facter) {
	p.input = inP
}

func (p *BaseFact) SetData(data interface{}) {
	p.data = data
}

func (p *BaseFact) SetAdditionalInputs(plugins []Facter) {
	p.additionalInputs = plugins
}

// Default implementations for support methods
func (p *BaseFact) SupportedConnections() (plugin.SupportLevel, []string) {
	return plugin.SupportNone, []string{}
}

func (p *BaseFact) SupportedInputFormats() (plugin.SupportLevel, []data.DataFormat) {
	return plugin.SupportNone, []data.DataFormat{}
}

func ValidateConnection(p Facter) error {
	connectionSupport, supportedConnections := p.SupportedConnections()

	log.WithFields(log.Fields{
		"fact":                  p.GetId(),
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
			p.SetConnection(connPlug)
			return nil
		}
	}
	return &plugin.ErrSupportNone{
		Plugin:        p.GetName(),
		SupportType:   "connection",
		SupportPlugin: connPlug.PluginName()}
}

func ValidateInput(p Facter) error {
	inputFormatSupport, supportedInputFormats := p.SupportedInputFormats()
	log.WithFields(log.Fields{
		"fact":             p.GetName(),
		"input-support":    inputFormatSupport,
		"supported-inputs": supportedInputFormats,
	}).Debug("validating input")

	if (inputFormatSupport == plugin.SupportOptional ||
		inputFormatSupport == plugin.SupportNone) &&
		len(supportedInputFormats) == 0 && p.GetInputName() == "" {
		return nil
	}

	if inputFormatSupport == plugin.SupportRequired && p.GetInputName() == "" {
		return &plugin.ErrSupportRequired{Plugin: p.GetName(), SupportType: "inputFormat"}
	}

	if p.GetInputName() != "" {
		inPlug := Facts[p.GetInputName()]
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
				Plugin: inPlug.GetName(), SupportType: "inputFormat"}
		}

		for _, s := range supportedInputFormats {
			if inPlug.GetFormat() == s {
				p.SetInput(inPlug)
				return nil
			}
		}

		return &plugin.ErrSupportNone{
			SupportType:   "inputFormat",
			SupportPlugin: string(inPlug.GetFormat()),
			Plugin:        p.GetName(),
		}
	}

	return &plugin.ErrSupportNotFound{
		SupportType:   "input",
		Plugin:        p.GetName(),
		SupportPlugin: p.GetInputName()}
}

func LoadAdditionalInputs(p Facter) []error {
	log.WithFields(log.Fields{"fact": p.GetId()}).
		Debug("loading additional inputs")

	if len(p.GetAdditionalInputNames()) == 0 {
		return nil
	}

	plugins := []Facter{}
	errs := []error{}
	for _, n := range p.GetAdditionalInputNames() {
		inPlug := Facts[n]
		if inPlug == nil {
			errs = append(errs, &plugin.ErrSupportNotFound{
				Plugin:        p.GetName(),
				SupportType:   "input",
				SupportPlugin: n,
			})
			continue
		}

		if inPlug.GetFormat() == "" {
			errs = append(errs, &plugin.ErrSupportRequired{
				Plugin:      inPlug.GetName(),
				SupportType: "inputFormat"})
			continue
		}

		plugins = append(plugins, inPlug)
	}

	if len(errs) > 0 {
		return errs
	}

	p.SetAdditionalInputs(plugins)
	return nil
}

// Collect is the main method for collecting data from the fact.
func (p *BaseFact) Collect() {}
