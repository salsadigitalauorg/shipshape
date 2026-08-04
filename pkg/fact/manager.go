package fact

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
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
		breach.TemplateFuncs["lookupFactAsStringMap"] = LookupFactAsStringMap

		m = &manager{
			Manager: pluginmanager.NewManager[Facter](),
		}
	}
	return m
}

// LookupFactAsStringMap looks up a fact by name and returns the string value
// for the given key. It is format-aware and never panics: an unknown input, an
// unsupported data format, or a missing key all return an empty string. This is
// registered as the "lookupFactAsStringMap" breach template function.
func LookupFactAsStringMap(inputName string, key string) string {
	input := Manager().FindPlugin(inputName)
	if input == nil {
		return ""
	}

	if input.GetFormat() == data.FormatMapString {
		m := data.AsMapString(input.GetData())
		if v, ok := m[key]; ok {
			return v
		}
	}

	return ""
}

func (m *manager) GetFactoriesKeys() []string {
	return plugin.GetFactoriesKeys[Facter](m.GetFactories())
}

// ResetPlugins clears the registered plugin instances, as well as the
// manager's own collected-facts cache. Overrides
// pluginmanager.Manager.ResetPlugins(), which only clears plugin instances:
// without also clearing m.collected here, a fact name collected by one test
// (or run) would be silently skipped by CollectFact's de-dup check
// (utils.StringSliceContains(m.collected, name)) in a later test/run that
// reuses the same name with a fresh plugin instance, even though that fresh
// instance was never actually collected. Production code never calls
// ResetPlugins mid-process, so this only affects test isolation - but it
// affects it significantly, since fact names like "primary"/"current" are
// reused across many _test.go files sharing the package-level manager
// singleton.
func (m *manager) ResetPlugins() {
	m.Manager.ResetPlugins()
	m.collected = nil
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

	var inputErrored bool
	if f.GetInputName() != "" {
		log.WithField("fact", name).
			WithField("inputName", f.GetInputName()).
			Debug("collect input")
		inputF := m.FindPlugin(f.GetInputName())
		m.CollectFact(f.GetInputName(), inputF)
		if inputF != nil && len(inputF.GetErrors()) > 0 {
			inputErrored = true
		}
	}

	if len(f.GetAdditionalInputNames()) > 0 {
		for _, n := range f.GetAdditionalInputNames() {
			log.WithField("fact", name).
				WithField("additionalInputName", n).
				Debug("collect additional input")
			additionalInputF := m.FindPlugin(n)
			m.CollectFact(n, additionalInputF)
			if additionalInputF != nil && len(additionalInputF.GetErrors()) > 0 {
				inputErrored = true
			}
		}
	}

	if inputErrored {
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
