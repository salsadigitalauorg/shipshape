package file

import (
	"errors"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

type ReadMultiple struct {
	fact.BaseFact

	// Plugin fields.
	Files []string `yaml:"files"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=ReadMultiple --package=file

func init() {
	fact.GetManager().Register("file:read:multiple", func(n string) fact.Facter {
		return NewReadMultiple(n)
	})
}

func NewReadMultiple(id string) *ReadMultiple {
	return &ReadMultiple{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *ReadMultiple) GetName() string {
	return "file:read:multiple"
}

func (p *ReadMultiple) SupportedInputFormats() (plugin.SupportLevel, []data.DataFormat) {
	return plugin.SupportOptional, []data.DataFormat{data.FormatMapString}
}

func (p *ReadMultiple) Collect() {
	log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
		"project-dir": config.ProjectDir,
	}).Info("collecting files data")

	if p.GetInput() == nil && len(p.Files) == 0 {
		p.AddErrors(errors.New("no files specified"))
		return
	}

	if p.GetInput() != nil {
		switch p.GetInput().GetFormat() {
		case data.FormatMapString:
			p.Format = data.FormatMapBytes
			res := map[string][]byte{}
			filenameMap := data.AsMapString(p.GetInput().GetData())
			for k, filename := range filenameMap {
				fullpath := filepath.Join(config.ProjectDir, filename)
				if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
					p.AddErrors(err)
					continue
				}
				fData, err := os.ReadFile(fullpath)
				if err != nil {
					p.AddErrors(err)
					continue
				}
				res[k] = fData
			}
			p.SetData(res)
			return
		}
	}

	p.Format = data.FormatMapBytes
	res := map[string][]byte{}
	for _, filename := range p.Files {
		fullpath := filepath.Join(config.ProjectDir, filename)
		if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
			p.AddErrors(err)
			continue
		}

		fData, err := os.ReadFile(fullpath)
		if err != nil {
			p.AddErrors(err)
			continue
		}
		res[filename] = fData
	}
	p.SetData(res)
}
