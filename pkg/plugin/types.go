// Package plugin provides the base interfaces and types for all Shipshape plugins.
package plugin

import "fmt"

// Plugin is the base interface that all plugins must implement.
type Plugin interface {
	// GetName returns the unique identifier for this plugin type.
	GetName() string
	// GetId returns the unique identifier for this plugin instance.
	GetId() string
}

// Registry represents a generic plugin registry.
type Registry[T Plugin] map[string]func(string) T

// RegistryNoId represents a plugin registry for plugins that don't require ids.
type RegistryNoId[T Plugin] map[string]func() T

// SupportLevel defines the level of support for plugin dependencies.
type SupportLevel string

const (
	SupportRequired SupportLevel = "required"
	SupportOptional SupportLevel = "optional"
	SupportNone     SupportLevel = "not-supported"
)

// ErrSupportRequired is returned when a required plugin dependency is missing.
type ErrSupportRequired struct {
	Plugin      string
	SupportType string
}

func (m *ErrSupportRequired) Error() string {
	return fmt.Sprintf("%s required for '%s'", m.SupportType, m.Plugin)
}

// ErrSupportNotFound is returned when a plugin dependency cannot be found.
type ErrSupportNotFound struct {
	Plugin        string
	SupportType   string
	SupportPlugin string
}

func (m *ErrSupportNotFound) Error() string {
	return fmt.Sprintf("%s '%s' not found for '%s'",
		m.SupportType, m.SupportPlugin, m.Plugin)
}

// ErrSupportNone is returned when a plugin dependency is not supported.
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
