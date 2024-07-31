package docker

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/docker"
	"github.com/salsadigitalauorg/shipshape/pkg/env"
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
	NoTag    bool   `yaml:"no-tag"`
	ArgsFrom string `yaml:"args-from"`
	// Ignore is a list of Docker images to ignore.
	// Env vars can be provided and will be resolved
	// against args if ArgsFrom is set.
	Ignore []string `yaml:"ignore"`
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

func (p *Images) resolveIgnore(envMapMap map[string]map[string]string) {
	if p.ArgsFrom == "" || envMapMap == nil {
		return
	}

	newIgnore := []string{}

	for _, i := range p.Ignore {
		resI := i
		for _, envMap := range envMapMap {
			var err error
			resI, err = env.ResolveValue(envMap, i)
			if err != nil {
				log.WithFields(log.Fields{
					"fact-plugin": p.PluginName(),
					"fact":        p.Name,
					"error":       err,
				}).Error("could not resolve ignore value")
				p.errors = append(p.errors, err)
				return
			}
			if resI != i {
				break
			}
		}
		newIgnore = append(newIgnore, resI)
	}
	p.Ignore = newIgnore
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
		p.errors = append(p.errors, &fact.ErrSupportNone{
			Plugin:        p.Name,
			SupportType:   "input data format",
			SupportPlugin: string(p.input.GetFormat())})
	}

	if fileBytesMap == nil {
		return
	}

	envMap := map[string]map[string]string{}
	if p.ArgsFrom != "" {
		if p.additionalInputs == nil {
			p.errors = append(p.errors, &fact.ErrSupportRequired{
				Plugin: p.Name, SupportType: "additional inputs"})
			return
		}

		for _, i := range p.additionalInputs {
			if i.GetName() == p.ArgsFrom {
				envMap = data.AsMapNestedString(i.GetData())
				break
			}
		}
	}

	p.resolveIgnore(envMap)

	baseImagesMap := map[string][]string{}
	for fn, fBytes := range fileBytesMap {
		baseImages, err := docker.Parse(fBytes, envMap[fn], p.NoTag, p.Ignore)
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
