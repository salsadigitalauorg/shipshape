package yaml

import (
	"errors"
	"fmt"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
	"gopkg.in/yaml.v3"
)

type Lookup struct {
	// Common fields.
	Name           string          `yaml:"name"`
	Format         fact.FactFormat `yaml:"format"`
	ConnectionName string          `yaml:"connection"`
	InputName      string          `yaml:"input"`
	connection     connection.Connectioner
	input          fact.Facter
	errors         []error
	data           interface{}

	// Plugin fields.
	Path  string `yaml:"path"`
	Regex string `yaml:"regex"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Lookup --package=yaml

func init() {
	fact.Registry["yaml.lookup"] = func(n string) fact.Facter { return &Lookup{Name: n} }
}

func (p *Lookup) PluginName() string {
	return "yaml.lookup"
}

func (p *Lookup) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Lookup) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportRequired, []string{"file.lookup"}
}

func (p *Lookup) Gather() {
	inputData := p.input.GetData()
	if inputData == nil {
		return
	}

	filesData := inputData.(map[string][]byte)
	if len(filesData) == 0 {
		return
	}

	data := map[string]map[string]string{}
	for f, fBytes := range filesData {
		n := yaml.Node{}
		err := yaml.Unmarshal(fBytes, &n)
		if err != nil {
			p.errors = append(p.errors, err)
			continue
		}

		data[f] = map[string]string{}
		foundNodes, err := utils.LookupYamlPath(&n, p.Path)
		if err != nil {
			p.errors = append(p.errors, err)
			continue
		}

		for _, node := range foundNodes {
			fmt.Printf("node: %#v\n", node)
			data[f][node.Anchor] = node.Value
		}

	}

	switch p.Format {
	case fact.FormatMapStringList:
		p.data = data
	default:
		p.errors = append(p.errors, errors.New("unsupported format"))
	}
}
