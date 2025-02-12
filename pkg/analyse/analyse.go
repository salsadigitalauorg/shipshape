package analyse

import (
	"sort"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

var Registry = map[string]func(string) Analyser{}
var Analysers = map[string]Analyser{}
var Errors = []error{}

func RegistryKeys() []string {
	keys := []string{}
	for k := range Registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ParseConfig parses the raw config and creates the analysers.
func ParseConfig(raw map[string]map[string]interface{}) {
	count := 0
	log.WithField("registry", RegistryKeys()).Debug("analysers")
	for name, pluginConf := range raw {
		for pluginName, pluginMap := range pluginConf {
			f, ok := Registry[pluginName]
			if !ok {
				continue
			}

			// Convert the map to yaml, then parse it into the plugin.
			// Not catching any errors here since the yaml content is known.
			pluginYaml, _ := yaml.Marshal(pluginMap)
			p := f(name)
			yaml.Unmarshal(pluginYaml, p)

			log.WithFields(log.Fields{
				"check":       name,
				"plugin":      pluginName,
				"description": p.GetDescription(),
				"input":       p.GetInputName(),
			}).Debug("parsed analyser")
			Analysers[name] = p
			count++
		}
	}
	log.Infof("parsed %d analysers", count)
}

func ValidateInputs() {
	for _, p := range Analysers {
		if err := p.ValidateInput(); err != nil {
			Errors = append(Errors, err)
		}
	}
}

func AnalyseAll() map[string]result.Result {
	results := map[string]result.Result{}
	for _, p := range Analysers {
		if p.PreProcessInput() {
			p.Analyse()
		}
		result := p.GetResult()
		results[p.GetId()] = result
		log.WithField("analyser", p.GetId()).WithFields(result.LogFields()).
			Debug("analysed result")
	}
	return results
}
