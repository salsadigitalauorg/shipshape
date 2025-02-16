package analyse

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
)

type RegexNotMatch struct {
	BaseAnalyser `yaml:",inline"`
	Pattern      string `yaml:"pattern"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=RegexNotMatch --package=analyse

func init() {
	Manager().RegisterFactory("regex:not-match", func(id string) Analyser { return NewRegexNotMatch(id) })
}

func (p *RegexNotMatch) GetName() string {
	return "regex:not-match"
}

func (p *RegexNotMatch) Analyse() {
	contextLogger := log.WithFields(log.Fields{
		"plugin": p.GetName(),
		"id":     p.GetId(),
	})

	contextLogger.WithFields(log.Fields{
		"input":        p.GetInputName(),
		"input-format": p.GetInput().GetFormat(),
	}).Debug("analysing")

	switch p.input.GetFormat() {

	case data.FormatNil:
		breach.EvaluateTemplate(p, &breach.ValueBreach{
			Value: fmt.Sprintf("%s is nil", p.GetInputName()),
		}, nil)
		return

	case data.FormatMapNestedString:
		inputData := data.AsMapNestedString(p.input.GetData())
		for k, kvs := range inputData {
			for k2, v := range kvs {
				match, _ := regexp.MatchString(p.Pattern, v)
				if !match {
					breach.EvaluateTemplate(p, &breach.KeyValueBreach{
						Key:        k,
						ValueLabel: k2,
						Value:      v,
					}, p.Remediation)
					continue
				}
			}
		}

	case data.FormatString:
		inputData := data.AsString(p.input.GetData())
		match, _ := regexp.MatchString(p.Pattern, inputData)
		if !match {
			breach.EvaluateTemplate(p, &breach.ValueBreach{
				Value: fmt.Sprintf("%s equals '%s'", p.InputName, inputData),
			}, p.Remediation)
		}

	default:
		log.WithField("input-format", p.input.GetFormat()).Debug("unsupported input format")
		breach.EvaluateTemplate(p, &breach.ValueBreach{
			Value: fmt.Sprintf("unsupported input format %s", p.input.GetFormat()),
		}, nil)
	}
}
