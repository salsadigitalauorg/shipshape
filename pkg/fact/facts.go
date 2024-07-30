package fact

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

var Registry = map[string]func(string) Facter{}

// OnlyFactNames is a list of fact names to collect.
// If empty, all facts are collected.
var OnlyFactNames = []string{}

var Facts = map[string]Facter{}
var Errors = []error{}
var collected = []string{}

func init() {
	breach.TemplateFuncs["lookupFactAsStringMap"] = func(inputName string, key string) string {
		input := GetInstance(inputName)
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
}

func GetInstance(name string) Facter {
	if p, ok := Facts[name]; !ok {
		return nil
	} else {
		return p
	}
}

func ParseConfig(raw map[string]map[string]interface{}) {
	count := 0
	log.WithField("registry", Registry).Debug("available facts")
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

			log.WithField("fact", fmt.Sprintf("%#v", p)).Debug("parsed fact")
			Facts[name] = p
			count++
		}
	}
	log.Infof("parsed %d facts", count)
}

func CollectAllFacts() {
	for name, p := range Facts {
		if len(OnlyFactNames) > 0 &&
			!utils.StringSliceContains(OnlyFactNames, name) {
			continue
		}
		CollectFact(name, p)
	}
}

func CollectFact(name string, f Facter) {
	log.WithField("fact", f).Debug("collecting fact")
	if f.GetInputName() != "" {
		CollectFact(f.GetInputName(), GetInstance(f.GetInputName()))
	}

	if utils.StringSliceContains(collected, name) {
		return
	}

	if err := f.ValidateConnection(); err != nil {
		Errors = append(Errors, err)
		log.WithField("fact", name).WithError(err).
			Error("failed to validate connection")
		return
	}

	if err := f.ValidateInput(); err != nil {
		Errors = append(Errors, err)
		log.WithField("fact", name).WithError(err).
			Error("failed to validate input")
		return
	}

	log.WithField("fact", name).Infof("collecting fact")
	f.Collect()
	if len(f.GetErrors()) > 0 {
		Errors = append(Errors, f.GetErrors()...)
	}

	log.WithFields(log.Fields{
		"fact": name,
		"data": f.GetData(),
	}).Trace("collected fact")
	collected = append(collected, name)
}
