package connection

type Connectioner interface {
	PluginName() string
	GetName() string
	Run() ([]byte, error)
}
