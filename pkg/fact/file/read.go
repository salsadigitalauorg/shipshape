package file

import (
	"errors"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

type Read struct {
	fact.BaseFact

	// Plugin fields.
	Path string `yaml:"path"`
}

func init() {
	fact.GetManager().Register("file:read", func(n string) fact.Facter {
		return NewRead(n)
	})
}

func NewRead(id string) *Read {
	return &Read{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *Read) GetName() string {
	return "file:read"
}

func (p *Read) SupportedConnections() (plugin.SupportLevel, []string) {
	return plugin.SupportNone, nil
}

func (p *Read) SupportedInputs() (plugin.SupportLevel, []string) {
	return plugin.SupportNone, nil
}

func (p *Read) Collect() {
	log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
		"project-dir": config.ProjectDir,
		"path":        p.Path,
	}).Info("verifying file existence")

	fullpath := filepath.Join(config.ProjectDir, p.Path)
	if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
		p.AddErrors(err)
		return
	}

	fData, err := os.ReadFile(fullpath)
	if err != nil {
		p.AddErrors(err)
		return
	}
	p.SetData(fData)
}
