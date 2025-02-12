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
		return nil, &ErrSupportRequired{
			Plugin: p.GetName(), SupportType: "connection"}
	}

	plugin := connection.GetInstance(p.GetConnectionName())
	if plugin == nil {
		return nil, &ErrSupportNotFound{
			Plugin:        p.GetName(),
			SupportType:   "connection",
			SupportPlugin: p.GetConnectionName()}
	}

	for _, s := range supportedConnections {
		if plugin.PluginName() == s {
			return plugin, nil
		}
	}
	return nil, &ErrSupportNone{
		Plugin:        p.PluginName(),
		SupportType:   "connection",
		SupportPlugin: plugin.PluginName()}
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
		return nil, &ErrSupportRequired{Plugin: p.GetName(), SupportType: "input"}
	}

	if p.GetInputName() != "" {
		plugin := GetInstance(p.GetInputName())
		if plugin == nil {
			return nil, &ErrSupportNotFound{
				Plugin:        p.GetName(),
				SupportType:   "input",
				SupportPlugin: p.GetInputName()}
		}

		log.WithFields(log.Fields{
			"fact":                p.GetName(),
			"input-plugin":        plugin.GetName(),
			"input-plugin-format": plugin.GetFormat(),
		}).Debug("found input plugin")

		if plugin.GetFormat() == "" {
			return nil, &ErrSupportRequired{
				Plugin: plugin.GetName(), SupportType: "input data format"}
		}

		for _, s := range supportedInputs {
			if plugin.PluginName() == s {
				return plugin, nil
			}
		}

		return nil, &ErrSupportNone{
			SupportType:   "input",
			SupportPlugin: plugin.PluginName(),
			Plugin:        p.PluginName(),
		}
	}

	return nil, &ErrSupportNotFound{
		SupportType:   "input",
		Plugin:        p.GetName(),
		SupportPlugin: p.GetInputName()}
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
			errs = append(errs, &ErrSupportNotFound{
				Plugin:        p.GetName(),
				SupportType:   "additional input",
				SupportPlugin: n,
			})
			continue
		}

		if plugin.GetFormat() == "" {
			errs = append(errs, &ErrSupportRequired{
				Plugin:      plugin.GetName(),
				SupportType: "additional input data format"})
			continue
		}

		plugins = append(plugins, plugin)
	}

	return plugins, errs
}
