package pluginmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
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
	err := manager.RegisterFactory("testFactory", func(name string) *TestPlugin {
		return &TestPlugin{BasePlugin: plugin.BasePlugin{Id: name}}
	})
	assert.NoError(err)

	// Test duplicate registration
	err = manager.RegisterFactory("testFactory", func(name string) *TestPlugin {
		return &TestPlugin{BasePlugin: plugin.BasePlugin{Id: name}}
	})
	assert.Error(err)
	assert.Contains(err.Error(), "already registered")

	// Test getting plugin
	plugin, err := manager.GetPlugin("testFactory", "test")
	assert.NoError(err)
	assert.NotNil(plugin)
	assert.Equal("test-plugin", plugin.GetName())

	// Test getting non-existent plugin
	plugin, err = manager.GetPlugin("non-existent", "test")
	assert.Error(err)
	assert.Contains(err.Error(), "not found in registry")
	assert.Nil(plugin)

	// Test listing plugins
	plugins := manager.ListPlugins()
	assert.Equal(1, len(plugins))
	assert.Equal("testFactory", plugins[0])

	// Test reset
	manager.Reset()
	assert.Empty(manager.ListPlugins())
}
