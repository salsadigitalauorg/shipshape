package connection

import "github.com/salsadigitalauorg/shipshape/pkg/plugin"

type Connectioner interface {
	plugin.Plugin
	Run() ([]byte, error)
}
