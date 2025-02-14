package pluginmanager

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

// Manager handles plugin registration, validation, and lifecycle management.
type Manager[T plugin.Plugin] struct {
	mu       sync.RWMutex
	registry plugin.Registry[T]
	plugins  map[string]T
}

// NewManager creates a new plugin manager instance.
func NewManager[T plugin.Plugin]() *Manager[T] {
	return &Manager[T]{
		registry: make(plugin.Registry[T]),
		plugins:  make(map[string]T),
	}
}

// Register adds a new plugin to the registry.
func (m *Manager[T]) Register(name string, factory func(string) T) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.registry[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	log.WithField("plugin", name).Debug("registering plugin")
	m.registry[name] = factory
	return nil
}

// GetPlugin returns a plugin instance by name, creating it if it doesn't exist.
func (m *Manager[T]) GetPlugin(name string) (T, error) {
	m.mu.RLock()
	if plugin, exists := m.plugins[name]; exists {
		m.mu.RUnlock()
		return plugin, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if plugin, exists := m.plugins[name]; exists {
		return plugin, nil
	}

	factory, exists := m.registry[name]
	if !exists {
		var zero T
		return zero, fmt.Errorf("plugin '%s' not found in registry", name)
	}

	plugin := factory(name)
	m.plugins[name] = plugin
	return plugin, nil
}

// GetRegistry returns the plugin registry.
func (m *Manager[T]) GetRegistry() plugin.Registry[T] {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.registry
}

// ListPlugins returns a sorted list of registered plugin names.
func (m *Manager[T]) ListPlugins() []string {
	return plugin.GetRegistryKeys[T](m.registry)
}

// Reset clears all registered plugins and instances.
func (m *Manager[T]) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registry = make(plugin.Registry[T])
	m.plugins = make(map[string]T)
}

// Create a manager for fact plugins
// var factManager = NewManager[fact.Facter]()
