package pluginmanager

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

// Manager handles plugin registration, validation, and lifecycle management.
type Manager[T plugin.Plugin] struct {
	mu        sync.RWMutex
	factories plugin.Factories[T]
	plugins   map[string]T
	errors    []error
}

// NewManager creates a new plugin manager instance.
func NewManager[T plugin.Plugin]() *Manager[T] {
	return &Manager[T]{
		factories: make(plugin.Factories[T]),
		plugins:   make(map[string]T),
	}
}

// RegisterFactory adds a new plugin factory to the registry.
func (m *Manager[T]) RegisterFactory(name string, factory func(string) T) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.factories[name]; exists {
		return fmt.Errorf("plugin factory '%s' is already registered", name)
	}

	log.WithField("plugin", name).Debug("registering plugin factory")
	m.factories[name] = factory
	return nil
}

// GetFactories returns the plugin factories.
func (m *Manager[T]) GetFactories() plugin.Factories[T] {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.factories
}

// FindPlugin returns a plugin instance by name.
func (m *Manager[T]) FindPlugin(name string) T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[name]
}

// GetPlugin returns a plugin instance by plugin name and id,
// creating it if it doesn't exist.
func (m *Manager[T]) GetPlugin(name string, id string) (T, error) {
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

	factory, exists := m.factories[name]
	if !exists {
		var zero T
		return zero, fmt.Errorf("plugin factory '%s' not found in registry", name)
	}

	plugin := factory(id)
	m.plugins[id] = plugin
	return plugin, nil
}

// GetPlugins returns the plugin instances.
func (m *Manager[T]) GetPlugins() map[string]T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins
}

// SetPlugins sets the plugin instances.
func (m *Manager[T]) SetPlugins(plugins map[string]T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins = plugins
}

// ListPlugins returns a sorted list of registered plugin names.
func (m *Manager[T]) ListPlugins() []string {
	return plugin.GetFactoriesKeys[T](m.factories)
}

// ResetPlugins resets the plugin instances.
func (m *Manager[T]) ResetPlugins() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins = make(map[string]T)
}

// AddErrors adds errors to the manager.
func (m *Manager[T]) AddErrors(errs ...error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, errs...)
}

// GetErrors returns the errors.
func (m *Manager[T]) GetErrors() []error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errors
}

// ResetErrors resets the errors.
func (m *Manager[T]) ResetErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = []error{}
}

// Reset clears all registered plugins and instances.
func (m *Manager[T]) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.factories = make(plugin.Factories[T])
	m.plugins = make(map[string]T)
	m.errors = []error{}
}

// You can use it like this:
// Create a manager for fact plugins
// factManager := pluginmanager.NewManager[fact.Facter]()

// // Register a plugin
// factManager.Register("command", func(name string) fact.Facter {
//     return &command.Command{Name: name}
// })

// // Get a plugin instance
// cmdPlugin, err := factManager.GetPlugin("command")
// if err != nil {
//     log.Fatal(err)
// }

// // List all registered plugins
// plugins := factManager.ListPlugins()
