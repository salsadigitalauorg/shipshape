package fact

import "github.com/salsadigitalauorg/shipshape/pkg/data"

type Facter interface {
	// Common plugin methods.
	PluginName() string
	GetName() string

	// Fact methods.
	GetErrors() []error
	GetConnectionName() string
	GetInputName() string
	GetAdditionalInputNames() []string
	GetData() interface{}
	GetFormat() data.DataFormat
	SupportedConnections() (SupportLevel, []string)
	ValidateConnection() error
	SupportedInputs() (SupportLevel, []string)
	ValidateInput() error
	LoadAdditionalInputs() []error
	Collect()
}

type SupportLevel string

const (
	SupportRequired SupportLevel = "required"
	SupportOptional SupportLevel = "optional"
	SupportNone     SupportLevel = "not-supported"
)

type ErrSupportRequired struct{ SupportType string }

func (m *ErrSupportRequired) Error() string {
	return m.SupportType + " is required"
}

type ErrSupportNotFound struct{ SupportType string }

func (m *ErrSupportNotFound) Error() string {
	return m.SupportType + " not found"
}

type ErrSupportNone struct{ SupportType string }

func (m *ErrSupportNone) Error() string {
	return m.SupportType + " not supported"
}
