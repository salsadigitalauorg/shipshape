package file

import (
	"errors"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
)

type ReadMultiple struct {
	// Common fields.
	Name           string          `yaml:"name"`
	Format         data.DataFormat `yaml:"format"`
	ConnectionName string          `yaml:"connection"`
	InputName      string          `yaml:"input"`
	connection     connection.Connectioner
	input          fact.Facter
	errors         []error
	data           interface{}

	// Plugin fields.
	Files []string `yaml:"files"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=ReadMultiple --package=file

func init() {
	fact.Registry["file:read:multiple"] = func(n string) fact.Facter {
		return &ReadMultiple{Name: n, Format: data.FormatMapBytes}
	}
}

func (p *ReadMultiple) PluginName() string {
	return "file:read:multiple"
}

func (p *ReadMultiple) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportNone, nil
}

func (p *ReadMultiple) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportOptional, []string{"yaml:key"}
}

func (p *ReadMultiple) Collect() {
	log.WithFields(log.Fields{
		"fact":        p.Name,
		"project-dir": config.ProjectDir,
	}).Info("collecting files data")

	if p.input == nil && len(p.Files) == 0 {
		p.errors = append(p.errors, errors.New("no files specified"))
		return
	}

	if p.input != nil {
		switch p.input.GetFormat() {
		case data.FormatMapString:
			p.Format = data.FormatMapBytes
			res := map[string][]byte{}
			filenameMap := data.AsMapString(p.input.GetData())
			for k, filename := range filenameMap {
				fullpath := filepath.Join(config.ProjectDir, filename)
				if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
					p.errors = append(p.errors, err)
					continue
				}
				fData, err := os.ReadFile(fullpath)
				if err != nil {
					p.errors = append(p.errors, err)
					continue
				}
				res[k] = fData
			}
			p.data = res
			return
		}
	}

	p.Format = data.FormatMapBytes
	res := map[string][]byte{}
	for _, filename := range p.Files {
		fullpath := filepath.Join(config.ProjectDir, filename)
		if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
			p.errors = append(p.errors, err)
			continue
		}

		fData, err := os.ReadFile(fullpath)
		if err != nil {
			p.errors = append(p.errors, err)
			continue
		}
		res[filename] = fData
	}
	p.data = res
}
