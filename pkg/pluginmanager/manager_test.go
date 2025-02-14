package pluginmanager

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/stretchr/testify/assert"
)

// TestPlugin implements the Plugin interface for testing.
type TestPlugin struct {
	plugin.BasePlugin
}

func (p *TestPlugin) GetName() string {
	return "test-plugin"
}

func (p *TestPlugin) GetId() string {
	return p.Id
}

func TestManager(t *testing.T) {
	manager := NewManager[*TestPlugin]()
	assert := assert.New(t)

	// Test registration
	err := manager.Register("test", func(name string) *TestPlugin {
		return &TestPlugin{BasePlugin: plugin.BasePlugin{Id: name}}
	})
	assert.NoError(err)

	// Test duplicate registration
	err = manager.Register("test", func(name string) *TestPlugin {
		return &TestPlugin{BasePlugin: plugin.BasePlugin{Id: name}}
	})
	assert.Error(err)
	assert.Contains(err.Error(), "already registered")

	// Test getting plugin
	plugin, err := manager.GetPlugin("test")
	assert.NoError(err)
	assert.NotNil(plugin)
	assert.Equal("test-plugin", plugin.GetName())

	// Test getting non-existent plugin
	plugin, err = manager.GetPlugin("non-existent")
	assert.Error(err)
	assert.Contains(err.Error(), "not found in registry")
	assert.Nil(plugin)

	// Test listing plugins
	plugins := manager.ListPlugins()
	assert.Equal(1, len(plugins))
	assert.Equal("test", plugins[0])

	// Test reset
	manager.Reset()
	assert.Empty(manager.ListPlugins())
}
