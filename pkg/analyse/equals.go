package analyse

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
)

// Equals is an analyser that checks if a fact is equal to a value.
// If a map is provided as input, the key is used to look up the value.
type Equals struct {
	BaseAnalyser `yaml:",inline"`
	Value        string `yaml:"value"`
	Key          string `yaml:"key"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=Equals --package=analyse

func init() {
	Manager().RegisterFactory("equals", func(id string) Analyser { return NewEquals(id) })
}

func (p *Equals) GetName() string {
	return "equals"
}

func (p *Equals) Analyse() {
	log.WithFields(log.Fields{
		"plugin":       p.GetName(),
		"id":           p.GetId(),
		"input":        p.GetInputName(),
		"input-format": p.GetInput().GetFormat(),
	}).Debug("analysing")

	switch p.input.GetFormat() {
	case data.FormatString:
		inputData := data.AsString(p.input.GetData())
		if inputData == p.Value {
			breach.EvaluateTemplate(p, &breach.ValueBreach{
				Value: fmt.Sprintf("%s equals '%s'", p.InputName, inputData),
			}, p.Remediation)
		}
	case data.FormatMapString:
		inputData := data.AsMapString(p.input.GetData())
		if inputData[p.Key] == p.Value {
			breach.EvaluateTemplate(p, &breach.ValueBreach{
				Value: fmt.Sprintf("%s equals '%s'", p.InputName, inputData[p.Key]),
			}, p.Remediation)
		}
	default:
		log.WithField("input-format", p.input.GetFormat()).Error("unsupported input format")
	}
}
