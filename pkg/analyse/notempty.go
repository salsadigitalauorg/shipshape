package analyse

import (
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	log "github.com/sirupsen/logrus"
)

type NotEmpty struct {
	// Common fields.
	Id                    string `yaml:"name"`
	Description           string `yaml:"description"`
	InputName             string `yaml:"input"`
	Severity              string `yaml:"severity"`
	breach.BreachTemplate `yaml:"breach-format"`
	Result                result.Result
	input                 fact.Facter
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=NotEmpty --package=analyse

func init() {
	Registry["not-empty"] = func(id string) Analyser { return NewNotEmpty(id) }
}

func (p *NotEmpty) PluginName() string {
	return "not-empty"
}

func (p *NotEmpty) Analyse() {
	log.WithField("input-format", p.input.GetFormat()).Debug("analysing")
	switch p.input.GetFormat() {
	case data.FormatMapNestedString:
		inputData := data.AsMapNestedString(p.input.GetData())
		log.WithField("inputData", inputData).Debug("analysing")
		if len(inputData) == 0 {
			return
		}
		for k, kvs := range inputData {
			for subK, v := range kvs {
				breach.EvaluateTemplate(p, &breach.KeyValueBreach{
					Key:        k,
					ValueLabel: subK,
					Value:      v,
				})
			}
		}
	}
}
