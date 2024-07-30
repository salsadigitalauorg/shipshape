package fact

import (
	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
)

func ValidatePluginConnection(p Facter) (connection.Connectioner, error) {
	connectionSupport, supportedConnections := p.SupportedConnections()
	log.WithFields(log.Fields{
		"fact":                  p.GetName(),
		"connection-support":    connectionSupport,
		"supported-connections": supportedConnections,
	}).Debug("validating connection")

	if (connectionSupport == SupportOptional ||
		connectionSupport == SupportNone) &&
		len(supportedConnections) == 0 && p.GetConnectionName() == "" {
		return nil, nil
	}

	if connectionSupport == SupportRequired && p.GetConnectionName() == "" {
		return nil, &ErrSupportRequired{SupportType: "connection"}
	}

	plugin := connection.GetInstance(p.GetConnectionName())
	if plugin == nil {
		return nil, &ErrSupportNotFound{SupportType: "connection"}
	}

	for _, s := range supportedConnections {
		if plugin.PluginName() == s {
			return plugin, nil
		}
	}
	return nil, &ErrSupportNone{SupportType: "connection"}
}

func ValidatePluginInput(p Facter) (Facter, error) {
	inputSupport, supportedInputs := p.SupportedInputs()
	log.WithFields(log.Fields{
		"fact":             p.GetName(),
		"input-support":    inputSupport,
		"supported-inputs": supportedInputs,
	}).Debug("validating input")

	if (inputSupport == SupportOptional ||
		inputSupport == SupportNone) &&
		len(supportedInputs) == 0 && p.GetInputName() == "" {
		return nil, nil
	}

	if inputSupport == SupportRequired && p.GetInputName() == "" {
		return nil, &ErrSupportRequired{SupportType: "input"}
	}

	if p.GetInputName() != "" {
		plugin := GetInstance(p.GetInputName())
		if plugin == nil {
			return nil, &ErrSupportNotFound{SupportType: "input"}
		}

		if plugin.GetFormat() == "" {
			return nil, &ErrSupportRequired{SupportType: "input data format"}
		}

		for _, s := range supportedInputs {
			if plugin.PluginName() == s {
				return plugin, nil
			}
		}
	}

	return nil, &ErrSupportNone{SupportType: "input"}
}

func LoadPluginAdditionalInputs(p Facter) ([]Facter, []error) {
	log.WithFields(log.Fields{"fact": p.GetName()}).
		Debug("loading additional inputs")

	if len(p.GetAdditionalInputNames()) == 0 {
		return nil, nil
	}

	plugins := []Facter{}
	errs := []error{}
	for _, n := range p.GetAdditionalInputNames() {
		plugin := GetInstance(n)
		if plugin == nil {
			errs = append(errs, &ErrSupportNotFound{SupportType: "additional input"})
			continue
		}

		if plugin.GetFormat() == "" {
			errs = append(errs, &ErrSupportRequired{SupportType: "additional input data format"})
			continue
		}

		plugins = append(plugins, plugin)
	}

	return plugins, errs
}
