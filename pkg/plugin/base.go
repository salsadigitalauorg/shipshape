package plugin

// BasePlugin provides common fields and functionality for all plugins.
type BasePlugin struct {
	// Common fields found across plugins
	Id string `yaml:"-"`

	// Internal fields
	errors []error
}

// Base getter methods
func (p *BasePlugin) GetId() string {
	return p.Id
}

func (p *BasePlugin) GetErrors() []error {
	return p.errors
}

// AddError adds an error to the plugin's error list
func (p *BasePlugin) AddErrors(errs ...error) {
	p.errors = append(p.errors, errs...)
}
