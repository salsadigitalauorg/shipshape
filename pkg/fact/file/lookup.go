package file

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/connection"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type Lookup struct {
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
	Path           string   `yaml:"path"`
	FileNamesOnly  bool     `yaml:"file-names-only"`
	Pattern        string   `yaml:"pattern"`
	ExcludePattern string   `yaml:"exclude-pattern"`
	SkipDirs       []string `yaml:"skip-dirs"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --plugin=Lookup --package=file

func init() {
	fact.Registry["file.lookup"] = func(n string) fact.Facter {
		return &Lookup{Name: n, FileNamesOnly: true}
	}
}

func (p *Lookup) PluginName() string {
	return "file.lookup"
}

func (p *Lookup) SupportedConnections() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Lookup) SupportedInputs() (fact.SupportLevel, []string) {
	return fact.SupportNone, []string{}
}

func (p *Lookup) Collect() {
	log.WithFields(log.Fields{
		"fact":        p.Name,
		"project-dir": config.ProjectDir,
		"path":        p.Path,
		"pattern":     p.Pattern,
	}).Info("looking up files")
	files, err := utils.FindFiles(filepath.Join(config.ProjectDir, p.Path), p.Pattern, p.ExcludePattern, p.SkipDirs)
	if err != nil {
		p.errors = append(p.errors, err)
		return
	}

	if p.FileNamesOnly {
		p.Format = data.FormatListString
		p.data = files
		return
	}

	filesDataMap := map[string][]byte{}
	for _, f := range files {
		fData, err := os.ReadFile(f)
		if err != nil {
			p.errors = append(p.errors, err)
			continue
		}
		filesDataMap[f] = fData
	}
	p.Format = data.FormatMapBytes
	p.data = filesDataMap
}
