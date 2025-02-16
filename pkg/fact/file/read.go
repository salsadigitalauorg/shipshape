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

type Read struct {
	fact.BaseFact `yaml:",inline"`

	// Plugin fields.
	Path string `yaml:"path"`
}

func init() {
	fact.Manager().RegisterFactory("file:read", func(n string) fact.Facter {
		return NewRead(n)
	})
}

func NewRead(id string) *Read {
	return &Read{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
			Format: data.FormatRaw,
		},
	}
}

func (p *Read) GetName() string {
	return "file:read"
}

func (p *Read) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	contextLogger.WithFields(log.Fields{
		"project-dir": config.ProjectDir,
		"path":        p.Path,
	}).Debug("verifying file existence")

	fullpath := filepath.Join(config.ProjectDir, p.Path)
	if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
		contextLogger.WithError(err).Debug("file does not exist")
		p.AddErrors(err)
		return
	}

	fData, err := os.ReadFile(fullpath)
	if err != nil {
		contextLogger.WithError(err).Debug("error reading file")
		p.AddErrors(err)
		return
	}
	p.SetData(fData)
}
