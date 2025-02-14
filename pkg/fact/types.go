package fact

import (
	"github.com/salsadigitalauorg/shipshape/pkg/connection"
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
	GetConnection() connection.Connectioner
	SupportedConnections() (plugin.SupportLevel, []string)
	ValidateConnection() error

	// Input methods
	GetInputName() string
	GetInput() Facter
	GetAdditionalInputNames() []string
	GetAdditionalInputs() []Facter
	SetInputName(name string)
	SupportedInputs() (plugin.SupportLevel, []string)
	ValidateInput() error
	LoadAdditionalInputs() []error

	// Collection
	Collect()
}
