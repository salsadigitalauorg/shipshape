package connection

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/pluginmanager"
)

// manager handles connection plugin registration and lifecycle.
type manager struct {
	*pluginmanager.Manager[Connectioner]
}

var m *manager

// Manager returns the connection manager.
func Manager() *manager {
	if m == nil {
		m = &manager{
			Manager: pluginmanager.NewManager[Connectioner](),
		}
	}
	return m
}

func (m *manager) GetFactoriesKeys() []string {
	return plugin.GetFactoriesKeys[Connectioner](m.GetFactories())
}

// ParseConfig parses the raw config and creates the connections.
func (m *manager) ParseConfig(raw map[string]map[string]interface{}) error {
	count := 0
	log.WithField("registry", m.GetFactoriesKeys()).Debug("available connections")
	for id, pluginConf := range raw {
		for pluginName, pluginIf := range pluginConf {
			plugin, err := m.GetPlugin(pluginName, id)
			if err != nil {
				return err
			}

			// Convert the map to yaml, then parse it into the plugin.
			pluginYaml, _ := yaml.Marshal(pluginIf)
			err = yaml.Unmarshal(pluginYaml, plugin)
			if err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"id":     id,
				"plugin": fmt.Sprintf("%#v", plugin),
			}).Debug("parsed connection")
			count++
		}
	}
	log.Infof("parsed %d connections", count)
	return nil
}
