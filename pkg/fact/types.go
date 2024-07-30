package fact

import (
	"fmt"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
)

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

type ErrSupportRequired struct {
	Plugin      string
	SupportType string
}

func (m *ErrSupportRequired) Error() string {
	return fmt.Sprintf("%s required for '%s'", m.SupportType, m.Plugin)
}

type ErrSupportNotFound struct {
	Plugin        string
	SupportType   string
	SupportPlugin string
}

func (m *ErrSupportNotFound) Error() string {
	return fmt.Sprintf("%s '%s' not found for '%s'",
		m.SupportType, m.SupportPlugin, m.Plugin)
}

type ErrSupportNone struct {
	Plugin        string
	SupportType   string
	SupportPlugin string
}

func (m *ErrSupportNone) Error() string {
	if m.SupportPlugin == "" {
		return fmt.Sprintf("%s not supported for '%s'", m.SupportType, m.Plugin)
	}
	return fmt.Sprintf("%s '%s' not supported for '%s'",
		m.SupportType, m.SupportPlugin, m.Plugin)
}
