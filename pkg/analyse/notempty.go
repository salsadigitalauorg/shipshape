package analyse

import (
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	log "github.com/sirupsen/logrus"
)

type NotEmpty struct {
	BaseAnalyser `yaml:",inline"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=NotEmpty --package=analyse

func init() {
	Manager().RegisterFactory("not:empty", func(id string) Analyser { return NewNotEmpty(id) })
}

func (p *NotEmpty) GetName() string {
	return "not:empty"
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
				}, p.Remediation)
			}
		}
	}
}
