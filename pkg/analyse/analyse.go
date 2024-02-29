package analyse

import (
	"fmt"

	"github.com/salsadigitalauorg/shipshape/pkg/result"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var Registry = map[string]func(string) Analyser{}
var Analysers = map[string]Analyser{}
var Errors = []error{}

func ParseConfig(raw map[string]map[string]interface{}) {
	count := 0
	log.WithField("registry", Registry).Debug("analysers")
	for name, pluginConf := range raw {
		for pluginName, pluginMap := range pluginConf {
			f, ok := Registry[pluginName]
			if !ok {
				continue
			}

			p := f(name)
			pluginYaml, err := yaml.Marshal(pluginMap)
			if err != nil {
				panic(err)
			}

			err = yaml.Unmarshal(pluginYaml, p)
			if err != nil {
				panic(err)
			}

			log.WithField("analyser", fmt.Sprintf("%#v", p)).Debug("parsed analyser")
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
		p.Analyse()
		results[p.GetName()] = p.GetResult()
	}
	return results
}
