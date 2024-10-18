package output

import (
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type Outputter interface {
	AddFlags(*cobra.Command)
	Output(io.Writer) error
}

var Registry = map[string]func(*result.ResultList) Outputter{}
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
		p := o(rl)
		yaml.Unmarshal(pluginYaml, p)

		log.WithFields(log.Fields{"plugin": pluginName}).Debug("parsed outputter")
		Outputters[pluginName] = p
		count++
	}
	log.Infof("parsed %d outputters", count)
}

func OutputAll(w io.Writer) error {
	for _, p := range Outputters {
		if err := p.Output(w); err != nil {
			return err
		}
	}
	return nil
}
