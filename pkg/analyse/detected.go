package analyse

import (
	"sort"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	log "github.com/sirupsen/logrus"
)

// Detected is an analyser that breaches once per entry in a map-shaped
// input (typically file:fingerprint), whenever that entry is present.
// Like Drift, there is no toggle for the inverse polarity: any entry in
// the map is itself the assertion failure, so there is nothing to
// configure beyond input/description/severity.
//
// Map keys are sorted before breaches are emitted, because Go's map
// iteration order is randomised and this analyser is typically used with
// multi-label input (e.g. several detected frameworks) - without
// sorting, breach order (and therefore e2e assertions on it) would be
// non-deterministic.
type Detected struct {
	BaseAnalyser `yaml:",inline"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=Detected --package=analyse

func init() {
	Manager().RegisterFactory("detected", func(id string) Analyser { return NewDetected(id) })
}

func (p *Detected) GetName() string {
	return "detected"
}

func (p *Detected) Analyse() {
	log.WithField("input-format", p.input.GetFormat()).Debug("analysing")

	switch p.input.GetFormat() {
	case data.FormatMapString:
		inputData := data.AsMapString(p.input.GetData())
		if len(inputData) == 0 {
			return
		}

		keys := make([]string, 0, len(inputData))
		for k := range inputData {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			breach.EvaluateTemplate(p, &breach.KeyValueBreach{
				KeyLabel:   "framework",
				Key:        k,
				ValueLabel: "likelihood",
				Value:      inputData[k],
			}, p.Remediation)
		}
	default:
		log.WithField("input-format", p.input.GetFormat()).Error("unsupported input format")
	}
}
