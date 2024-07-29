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

type Read struct {
	// Common fields.
	Name           string          `yaml:"name"`
	Format         data.DataFormat `yaml:"format"`
	ConnectionName string          `yaml:"connection"`
	InputName      string          `yaml:"input"`
	InputNames     []string        `yaml:"inputs"`
	connection     connection.Connectioner
	input          fact.Facter
	inputs         []fact.Facter
	errors         []error
	data           interface{}

	// Plugin fields.
	Path string `yaml:"path"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Read --package=file

func init() {
	fact.Registry["file:read"] = func(n string) fact.Facter {
		return &Read{Name: n, Format: data.FormatRaw}
	}
}

func (p *Read) PluginName() string {
	return "file:read"
}

func (p *Read) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportNone, nil
}

func (p *Read) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportNone, nil
}

func (p *Read) Collect() {
	log.WithFields(log.Fields{
		"fact":        p.Name,
		"project-dir": config.ProjectDir,
		"path":        p.Path,
	}).Info("verifying file existence")

	fullpath := filepath.Join(config.ProjectDir, p.Path)
	if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
		p.errors = append(p.errors, err)
		return
	}

	fData, err := os.ReadFile(fullpath)
	if err != nil {
		p.errors = append(p.errors, err)
		return
	}
	p.data = fData
}
