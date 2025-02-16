package docker

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/docker"
	"github.com/salsadigitalauorg/shipshape/pkg/env"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

type Images struct {
	fact.BaseFact `yaml:",inline"`

	// Plugin fields.
	NoTag    bool   `yaml:"no-tag"`
	ArgsFrom string `yaml:"args-from"`
	// Ignore is a list of Docker images to ignore.
	// Env vars can be provided and will be resolved
	// against args if ArgsFrom is set.
	Ignore []string `yaml:"ignore"`
}

func init() {
	fact.Manager().RegisterFactory("docker:images", func(n string) fact.Facter {
		return NewImages(n)
	})
}

func NewImages(id string) *Images {
	return &Images{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *Images) GetName() string {
	return "docker:images"
}

func (p *Images) SupportedInputFormats() (plugin.SupportLevel, []data.DataFormat) {
	return plugin.SupportRequired, []data.DataFormat{data.FormatMapBytes}
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
					"fact-plugin": p.GetName(),
					"fact":        p.GetId(),
					"error":       err,
				}).Error("could not resolve ignore value")
				p.AddErrors(err)
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
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	contextLogger.WithFields(log.Fields{
		"input":             p.GetInputName(),
		"input-plugin":      p.GetInput().GetName(),
		"additional-inputs": p.GetAdditionalInputNames(),
		"no-tag":            p.NoTag,
		"ignore":            p.Ignore,
	}).Debug("collecting data")

	var fileBytesMap map[string][]byte

	switch p.GetInput().GetFormat() {
	case data.FormatMapBytes:
		inputData := data.AsMapBytes(p.GetInput().GetData())
		if inputData == nil {
			return
		}

		fileBytesMap = inputData
	default:
		p.AddErrors(&plugin.ErrSupportNone{
			Plugin:        p.GetName(),
			SupportType:   "inputFormat",
			SupportPlugin: string(p.GetInput().GetFormat())})
	}

	if fileBytesMap == nil {
		return
	}

	envMap := map[string]map[string]string{}
	if p.ArgsFrom != "" {
		contextLogger.WithField("argsFrom", p.ArgsFrom).Debug("resolving env")
		if len(p.GetAdditionalInputNames()) == 0 {
			p.AddErrors(&plugin.ErrSupportRequired{
				Plugin: p.GetName(), SupportType: "additionalInputs"})
			return
		}

		for _, i := range p.GetAdditionalInputs() {
			if i.GetId() == p.ArgsFrom {
				envMap = data.AsMapNestedString(i.GetData())
				break
			}
		}

		contextLogger.WithField("envMap", envMap).Debug("resolved env")
	}
	p.resolveIgnore(envMap)

	baseImagesMap := map[string][]string{}
	for fn, fBytes := range fileBytesMap {
		baseImages, err := docker.Parse(fBytes, envMap[fn], p.NoTag, p.Ignore)
		if err != nil {
			log.WithField("error", err).Error("could not parse Dockerfile")
			p.AddErrors(err)
			return
		}

		p.Format = data.FormatMapListString
		baseImagesMap[fn] = []string{}
		for _, bi := range baseImages {
			baseImagesMap[fn] = append(baseImagesMap[fn], bi.String())
		}

		p.SetData(baseImagesMap)
		log.WithFields(log.Fields{
			"fact":       p.GetId(),
			"baseImages": fmt.Sprintf("%+v", baseImagesMap),
		}).Debug("parsed Dockerfile")
	}
}
