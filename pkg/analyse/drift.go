package analyse

import (
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	log "github.com/sirupsen/logrus"
)

// Drift is an analyser that breaches once per input, whenever its input
// (typically file:drift) is non-empty. There is no toggle for the inverse
// polarity: drift always means breach, so unlike allowed:list or
// regex:match there is nothing to configure - the presence of any drift
// content is itself the assertion failure.
type Drift struct {
	BaseAnalyser `yaml:",inline"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=Drift --package=analyse

func init() {
	Manager().RegisterFactory("drift", func(id string) Analyser { return NewDrift(id) })
}

func (p *Drift) GetName() string {
	return "drift"
}

func (p *Drift) Analyse() {
	log.WithField("input-format", p.input.GetFormat()).Debug("analysing")

	switch p.input.GetFormat() {
	case data.FormatListString:
		inputData := data.AsListString(p.input.GetData())
		if len(inputData) == 0 {
			return
		}
		breach.EvaluateTemplate(p, &breach.KeyValuesBreach{
			KeyLabel:   "file",
			Key:        p.GetInputName(),
			ValueLabel: "drift",
			Values:     inputData,
		}, p.Remediation)
	default:
		log.WithField("input-format", p.input.GetFormat()).Error("unsupported input format")
	}
}
