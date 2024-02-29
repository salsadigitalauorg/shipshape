package fact

type FactFormat string

const (
	FormatRaw             FactFormat = "raw"
	FormatList            FactFormat = "list"
	FormatMapBytes        FactFormat = "map-bytes"
	FormatMapString       FactFormat = "map-string"
	FormatMapYamlNodes    FactFormat = "map-yaml-nodes"
	FormatMapNestedString FactFormat = "map-nested-string"
	FormatYaml            FactFormat = "yaml"
	FormatJson            FactFormat = "json"
)

type Facter interface {
	PluginName() string
	GetName() string
	GetData() interface{}
	GetFormat() FactFormat
	GetConnectionName() string
	GetInputName() string
	GetErrors() []error
	SupportedConnections() (SupportLevel, []string)
	ValidateConnection() error
	SupportedInputs() (SupportLevel, []string)
	ValidateInput() error
	Gather()
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
