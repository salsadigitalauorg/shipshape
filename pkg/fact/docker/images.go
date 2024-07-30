package docker

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/docker"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
)

type Images struct {
	// Common fields.
	Name                 string          `yaml:"name"`
	Format               data.DataFormat `yaml:"format"`
	ConnectionName       string          `yaml:"connection"`
	InputName            string          `yaml:"input"`
	AdditionalInputNames []string        `yaml:"additional-inputs"`
	connection           connection.Connectioner
	input                fact.Facter
	additionalInputs     []fact.Facter
	errors               []error
	data                 interface{}

	// Plugin fields.
	ArgsFrom string `yaml:"argsFrom"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Images --package=docker

func init() {
	fact.Registry["docker:images"] = func(n string) fact.Facter { return &Images{Name: n} }
}

func (p *Images) PluginName() string {
	return "docker:images"
}

func (p *Images) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Images) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportRequired, []string{
		"file:read",
		"file:lookup",
		"file:read:multiple",
	}
}

func (p *Images) Collect() {
	log.WithFields(log.Fields{
		"fact-plugin":  p.PluginName(),
		"fact":         p.Name,
		"input":        p.GetInputName(),
		"input-plugin": p.input.PluginName(),
	}).Debug("collecting data")

	var fileBytesMap map[string][]byte

	switch p.input.GetFormat() {
	case data.FormatMapBytes:
		inputData := data.AsMapBytes(p.input.GetData())
		if inputData == nil {
			return
		}

		fileBytesMap = inputData
	default:
		p.errors = append(p.errors, &fact.ErrSupportNone{SupportType: "input data format"})
	}

	if fileBytesMap == nil {
		return
	}

	envMap := map[string]map[string]string{}
	if p.ArgsFrom != "" {
		if p.additionalInputs == nil {
			p.errors = append(p.errors, &fact.ErrSupportRequired{SupportType: "additional inputs"})
			return
		}

		for _, i := range p.additionalInputs {
			if i.GetName() == p.ArgsFrom {
				envMap = data.AsMapNestedString(i.GetData())
				break
			}
		}
	}

	baseImagesMap := map[string][]string{}
	for fn, fBytes := range fileBytesMap {
		baseImages, err := docker.Parse(fBytes, envMap[fn])
		if err != nil {
			log.WithField("error", err).Error("could not parse Dockerfile")
			p.errors = append(p.errors, err)
			return
		}

		p.Format = data.FormatMapListString
		baseImagesMap[fn] = []string{}
		for _, bi := range baseImages {
			baseImagesMap[fn] = append(baseImagesMap[fn], bi.String())
		}

		p.data = baseImagesMap
		log.WithFields(log.Fields{
			"fact":       p.Name,
			"baseImages": fmt.Sprintf("%+v", baseImagesMap),
		}).Debug("parsed Dockerfile")
	}
}
