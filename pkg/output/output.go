package output

import (
	"io"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type Outputter interface {
	Output(*result.ResultList) ([]byte, error)
}

var Registry = map[string]func() Outputter{}
var Outputters = map[string]Outputter{}

func registryKeys() []string {
	keys := []string{}
	for k := range Registry {
		keys = append(keys, k)
	}
	return keys
}

func ParseConfig(raw map[string]interface{}, rl *result.ResultList) {
	count := 0
	log.WithField("registry", registryKeys()).Debug("outputters")
	for pluginName, pluginMap := range raw {
		o, ok := Registry[pluginName]
		if !ok {
			continue
		}

		// Convert the map to yaml, then parse it into the plugin.
		// Not catching any errors here since the yaml content is known.
		pluginYaml, _ := yaml.Marshal(pluginMap)
		p := o()
		yaml.Unmarshal(pluginYaml, p)

		log.WithFields(log.Fields{"plugin": pluginName}).Debug("parsed outputter")
		Outputters[pluginName] = p
		count++
	}
	log.Infof("parsed %d outputters", count)
}

func OutputAll(rl *result.ResultList, w io.Writer) error {
	for _, p := range Outputters {
		buf, err := p.Output(rl)
		if err != nil {
			return err
		}

		if _, err := w.Write(buf); err != nil {
			return err
		}
	}
	return nil
}
