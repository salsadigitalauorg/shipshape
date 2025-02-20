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
	SetConnection(connection.Connectioner)
	SupportedConnections() (plugin.SupportLevel, []string)

	// Input methods
	GetInputName() string
	GetInput() Facter
	GetAdditionalInputNames() []string
	GetAdditionalInputs() []Facter
	SetInputName(name string)
	SetInput(Facter)
	SupportedInputFormats() (plugin.SupportLevel, []data.DataFormat)
	SetAdditionalInputs([]Facter)

	// Collection
	Collect()
}
