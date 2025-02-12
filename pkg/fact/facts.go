package fact

import (
	"sort"

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

func RegistryKeys() []string {
	keys := []string{}
	for k := range Registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func ParseConfig(raw map[string]map[string]interface{}) {
	count := 0
	log.WithField("registry", RegistryKeys()).Debug("available fact plugins")
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

			log.WithFields(log.Fields{
				"fact":   p.GetName(),
				"plugin": pluginName,
			}).Debug("parsed fact config")
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
	log.WithField("fact", name).Debug("starting CollectFact process")
	var inputF Facter
	if f.GetInputName() != "" {
		log.WithField("fact", name).
			WithField("inputName", f.GetInputName()).
			Debug("collect input")
		inputF = GetInstance(f.GetInputName())
		CollectFact(f.GetInputName(), inputF)
	}

	if len(f.GetAdditionalInputNames()) > 0 {
		for _, n := range f.GetAdditionalInputNames() {
			log.WithField("fact", name).
				WithField("additionalInputName", n).
				Debug("collect additional input")
			CollectFact(n, GetInstance(n))
		}
	}

	if inputF != nil && len(inputF.GetErrors()) > 0 {
		return
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

	if errs := f.LoadAdditionalInputs(); len(errs) != 0 {
		Errors = append(Errors, errs...)
		log.WithField("fact", name).WithField("errors", errs).
			Error("failed to load additional input")
		return
	}

	log.WithField("fact", name).Info("collecting fact")
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
