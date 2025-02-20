package connection

import (
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

// BaseConnection provides common fields and functionality for connection plugins.
type BaseConnection struct {
	plugin.BasePlugin `yaml:",inline"`
	Name              string `yaml:"name"`
	errors            []error
	data              []byte
}

func (p *BaseConnection) GetName() string {
	return p.Name
}

func (p *BaseConnection) GetErrors() []error {
	return p.errors
}

func (p *BaseConnection) AddErrors(errs ...error) {
	p.errors = append(p.errors, errs...)
}
