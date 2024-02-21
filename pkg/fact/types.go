package fact

type FactFormat string

const (
	FormatRaw  FactFormat = "raw"
	FormatList FactFormat = "list"
	FormatYaml FactFormat = "yaml"
	FormatJson FactFormat = "json"
)

type Facter interface {
	PluginName() string
	GetName() string
	GetFormat() FactFormat
	GetConnection() string
	GetErrors() []error
	Gather()
}

var Facts = map[string]Facter{}
