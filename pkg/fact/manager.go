package fact

import (
	"github.com/salsadigitalauorg/shipshape/pkg/pluginmanager"
)

// Manager handles fact plugin registration and lifecycle.
type Manager struct {
	*pluginmanager.Manager[Facter]
}

var m *Manager

// GetManager returns the fact manager.
func GetManager() *Manager {
	if m == nil {
		m = &Manager{
			Manager: pluginmanager.NewManager[Facter](),
		}
	}
	return m
}
