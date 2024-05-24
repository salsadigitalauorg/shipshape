package fact

import (
	"fmt"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var Registry = map[string]func(string) Facter{}
var Facts = map[string]Facter{}
var Errors = []error{}
var collected = []string{}

func ParseConfig(raw map[string]map[string]interface{}) {
	count := 0
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
		CollectFact(name, p)
	}
}

func CollectFact(name string, f Facter) {
	if f.GetInputName() != "" {
		CollectFact(f.GetInputName(), GetInstance(f.GetInputName()))
	}

	if utils.StringSliceContains(collected, name) {
		return
	}

	if err := f.ValidateConnection(); err != nil {
		Errors = append(Errors, err)
		log.WithField("fact", name).WithError(err).Error("failed to validate connection")
		return
	}

	if err := f.ValidateInput(); err != nil {
		Errors = append(Errors, err)
		log.WithField("fact", name).WithError(err).Error("failed to validate input")
		return
	}

	log.WithField("fact", name).Infof("collecting fact")
	f.Collect()

	log.WithFields(log.Fields{
		"fact": name,
		"data": f.GetData(),
	}).Trace("collected fact")
	collected = append(collected, name)
}

func GetInstance(name string) Facter {
	if p, ok := Facts[name]; !ok {
		return nil
	} else {
		return p
	}
}
