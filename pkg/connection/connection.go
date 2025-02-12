package connection

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var Registry = map[string]func(string) Connectioner{}
var Connections = map[string]Connectioner{}
var Errors = []error{}

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
	log.WithField("registry", RegistryKeys()).Debug("available connections")
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

			log.WithField("connection", fmt.Sprintf("%#v", p)).Debug("parsed connection")
			Connections[name] = p
			count++
		}
	}
	log.Infof("parsed %d connections", count)
}

func GetInstance(name string) Connectioner {
	if c, ok := Connections[name]; !ok {
		return nil
	} else {
		return c
	}
}
