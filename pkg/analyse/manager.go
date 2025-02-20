package analyse

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/pluginmanager"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

// manager handles analyser plugin registration and lifecycle.
type manager struct {
	*pluginmanager.Manager[Analyser]
}

var m *manager

// Manager returns the analyser manager.
func Manager() *manager {
	if m == nil {
		m = &manager{
			Manager: pluginmanager.NewManager[Analyser](),
		}
	}
	return m
}

func (m *manager) GetFactoriesKeys() []string {
	return plugin.GetFactoriesKeys[Analyser](m.GetFactories())
}

// ParseConfig parses the raw config and creates the analysers.
func (m *manager) ParseConfig(raw map[string]map[string]interface{}) error {
	log.WithField("registry", m.ListPlugins()).Debug("analysers")
	count := 0
	for id, pluginConf := range raw {
		for pluginName, pluginIf := range pluginConf {
			plugin, err := m.GetPlugin(pluginName, id)
			if err != nil {
				return err
			}

			// Convert the map to yaml, then parse it into the plugin.
			// Not catching any errors when marshalling since the yaml content is known.
			pluginYaml, _ := yaml.Marshal(pluginIf)
			err = yaml.Unmarshal(pluginYaml, plugin)
			if err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"id":          id,
				"plugin":      plugin.GetName(),
				"description": plugin.GetDescription(),
				"input":       plugin.GetInputName(),
			}).Debug("parsed analyser")
			count++
		}
	}
	log.Infof("parsed %d analysers", count)
	return nil
}

func (m *manager) ValidateInputs() {
	for _, plugin := range m.GetPlugins() {
		if err := plugin.ValidateInput(); err != nil {
			m.AddErrors(err)
		}
	}
}

// AnalyseAll runs all registered analysers and returns their results.
func (m *manager) AnalyseAll() map[string]result.Result {
	results := make(map[string]result.Result)
	for _, plugin := range m.GetPlugins() {
		if plugin.PreProcessInput() {
			plugin.Analyse()
		}

		result := plugin.GetResult()
		results[plugin.GetId()] = result

		log.WithField("analyser", plugin.GetId()).
			WithFields(result.LogFields()).
			Debug("analysed result")
	}

	return results
}
