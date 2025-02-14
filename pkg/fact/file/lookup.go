package file

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type Lookup struct {
	fact.BaseFact

	// Plugin fields.
	Path           string   `yaml:"path"`
	FileNamesOnly  bool     `yaml:"file-names-only"`
	Pattern        string   `yaml:"pattern"`
	ExcludePattern string   `yaml:"exclude-pattern"`
	SkipDirs       []string `yaml:"skip-dirs"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --package=file

func init() {
	fact.GetManager().Register("file:lookup", func(n string) fact.Facter {
		return NewLookup(n)
	})
}

func NewLookup(id string) *Lookup {
	return &Lookup{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
		FileNamesOnly: true,
	}
}

func (p *Lookup) GetName() string {
	return "file:lookup"
}

func (p *Lookup) SupportedConnections() (plugin.SupportLevel, []string) {
	return plugin.SupportNone, []string{}
}

func (p *Lookup) SupportedInputs() (plugin.SupportLevel, []string) {
	return plugin.SupportNone, []string{}
}

func (p *Lookup) Collect() {
	log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
		"project-dir": config.ProjectDir,
		"path":        p.Path,
		"pattern":     p.Pattern,
	}).Info("looking up files")
	files, err := utils.FindFiles(filepath.Join(config.ProjectDir, p.Path), p.Pattern, p.ExcludePattern, p.SkipDirs)
	if err != nil {
		p.AddErrors(err)
		return
	}

	if p.FileNamesOnly {
		p.Format = data.FormatListString
		p.SetData(files)
		return
	}

	filesDataMap := map[string][]byte{}
	for _, f := range files {
		fData, err := os.ReadFile(f)
		if err != nil {
			p.AddErrors(err)
			continue
		}
		filesDataMap[f] = fData
	}
	p.Format = data.FormatMapBytes
	p.SetData(filesDataMap)
}
