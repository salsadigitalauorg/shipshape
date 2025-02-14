package fact

import (
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

// Facter defines the interface for fact plugins
type Facter interface {
	plugin.Plugin

	// Data methods
	GetData() interface{}
	GetFormat() data.DataFormat

	// Connection methods
	GetConnectionName() string
	SupportedConnections() (plugin.SupportLevel, []string)
	ValidateConnection() error

	// Input methods
	GetInputName() string
	GetAdditionalInputNames() []string
	SupportedInputs() (plugin.SupportLevel, []string)
	ValidateInput() error
	LoadAdditionalInputs() []error

	// Collection
	Collect()
}
