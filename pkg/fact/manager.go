package fact

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/pluginmanager"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

// manager handles fact plugin registration and lifecycle.
type manager struct {
	*pluginmanager.Manager[Facter]
	// collected is a list of fact names that have already been collected.
	collected []string
}

var m *manager

// OnlyFactNames is a list of fact names to collect.
// If empty, all facts are collected.
var OnlyFactNames = []string{}

// Manager returns the fact manager.
func Manager() *manager {
	if m == nil {
		// Add a template function to lookup a fact as a string map.
		breach.TemplateFuncs["lookupFactAsStringMap"] = func(inputName string, key string) string {
			input := Manager().FindPlugin(inputName)
			if input == nil {
				return ""
			}
			ifcMap := input.GetData().(map[string]interface{})
			val, ok := ifcMap[key]
			if !ok {
				return ""
			}
			return val.(string)
		}

		m = &manager{
			Manager: pluginmanager.NewManager[Facter](),
		}
	}
	return m
}

func (m *manager) GetFactoriesKeys() []string {
	return plugin.GetFactoriesKeys[Facter](m.GetFactories())
}

// ParseConfig parses the raw config and creates the facts.
func (m *manager) ParseConfig(raw map[string]map[string]interface{}) error {
	count := 0
	log.WithField("registry", m.GetFactoriesKeys()).Debug("available fact plugins")
	for id, pluginConf := range raw {
		for pluginName, pluginIf := range pluginConf {
			log.WithField("pluginIf", pluginIf).Trace("parsing fact config")
			p, err := Manager().GetPlugin(pluginName, id)
			if err != nil {
				return err
			}

			// Convert the map to yaml, then parse it into the plugin.
			// Not catching any errors when marshalling since the yaml content is known.
			pluginYaml, _ := yaml.Marshal(pluginIf)
			err = yaml.Unmarshal(pluginYaml, p)
			if err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"id":     p.GetId(),
				"plugin": pluginName,
			}).Debug("parsed fact")

			log.WithField("fact", fmt.Sprintf("%#v", p)).Trace("parsed fact")
			count++
		}
	}
	log.Infof("parsed %d facts", count)
	return nil
}

// CollectAllFacts collects all facts.
func (m *manager) CollectAllFacts() {
	for name, p := range m.GetPlugins() {
		if len(OnlyFactNames) > 0 &&
			!utils.StringSliceContains(OnlyFactNames, name) {
			continue
		}
		m.CollectFact(name, p)
	}
}

// CollectFact collects a fact.
func (m *manager) CollectFact(name string, f Facter) {
	log.WithField("fact", name).Debug("starting CollectFact process")
	var inputF Facter
	if f.GetInputName() != "" {
		log.WithField("fact", name).
			WithField("inputName", f.GetInputName()).
			Debug("collect input")
		inputF = m.FindPlugin(f.GetInputName())
		m.CollectFact(f.GetInputName(), inputF)
	}

	if len(f.GetAdditionalInputNames()) > 0 {
		for _, n := range f.GetAdditionalInputNames() {
			log.WithField("fact", name).
				WithField("additionalInputName", n).
				Debug("collect additional input")
			inputF = m.FindPlugin(n)
			m.CollectFact(n, inputF)
		}
	}

	if inputF != nil && len(inputF.GetErrors()) > 0 {
		return
	}

	if utils.StringSliceContains(m.collected, name) {
		return
	}

	if err := ValidateConnection(f); err != nil {
		m.AddErrors(err)
		log.WithField("fact", name).WithError(err).
			Error("failed to validate connection")
		return
	}

	if err := ValidateInput(f); err != nil {
		m.AddErrors(err)
		log.WithField("fact", name).WithError(err).
			Error("failed to validate input")
		return
	}

	if errs := LoadAdditionalInputs(f); len(errs) != 0 {
		m.AddErrors(errs...)
		log.WithField("fact", name).WithField("errors", errs).
			Error("failed to load additional input")
		return
	}

	log.WithField("fact", name).Info("collecting fact")
	f.Collect()
	if len(f.GetErrors()) > 0 {
		m.AddErrors(f.GetErrors()...)
	}

	log.WithFields(log.Fields{
		"fact": name,
		"data": f.GetData(),
	}).Trace("collected fact")
	m.collected = append(m.collected, name)
}
