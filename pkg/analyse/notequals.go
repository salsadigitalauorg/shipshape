package analyse

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
)

// NotEquals is an analyser that checks if a fact is not equal to a value.
// If a map is provided as input, the key is used to look up the value.
type NotEquals struct {
	BaseAnalyser `yaml:",inline"`
	Value        string `yaml:"value"`
	Key          string `yaml:"key"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=NotEquals --package=analyse

func init() {
	Manager().RegisterFactory("not:equals", func(id string) Analyser { return NewNotEquals(id) })
}

func (p *NotEquals) GetName() string {
	return "not:equals"
}

func (p *NotEquals) Analyse() {
	log.WithFields(log.Fields{
		"plugin":       p.GetName(),
		"id":           p.GetId(),
		"input":        p.GetInputName(),
		"input-format": p.GetInput().GetFormat(),
	}).Debug("analysing")

	switch p.input.GetFormat() {
	case data.FormatString:
		inputData := data.AsString(p.input.GetData())
		if inputData != p.Value {
			breach.EvaluateTemplate(p, &breach.ValueBreach{
				Value: fmt.Sprintf("%s does not equal '%s'", p.InputName, p.Value),
			}, p.Remediation)
		}
	case data.FormatMapString:
		inputData := data.AsMapString(p.input.GetData())
		if inputData[p.Key] != p.Value {
			breach.EvaluateTemplate(p, &breach.ValueBreach{
				Value: fmt.Sprintf("%s does not equal '%s'", p.InputName, p.Value),
			}, p.Remediation)
		}
	default:
		log.WithField("input-format", p.input.GetFormat()).Error("unsupported input format")
	}
}
