package analyse

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
)

type RegexMatch struct {
	BaseAnalyser `yaml:",inline"`
	Pattern      string `yaml:"pattern"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=RegexMatch --package=analyse

func init() {
	Manager().RegisterFactory("regex:match", func(id string) Analyser {
		return NewRegexMatch(id)
	})
}

func (p *RegexMatch) GetName() string {
	return "regex:match"
}

func (p *RegexMatch) Analyse() {
	input := p.GetInput()
	if input == nil {
		return
	}

	re, err := regexp.Compile(p.Pattern)
	if err != nil {
		p.AddErrors(err)
		return
	}

	switch input.GetFormat() {
	case data.FormatNil:
		return
	case data.FormatMapNestedString:
		inputData := data.AsMapNestedString(input.GetData())
		for k, kvs := range inputData {
			for k2, v := range kvs {
				if re.MatchString(v) {
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
		inputData := data.AsString(input.GetData())
		if re.MatchString(inputData) {
			breach.EvaluateTemplate(p, &breach.ValueBreach{
				Value: fmt.Sprintf("%s equals '%s'", p.InputName, inputData),
			}, p.Remediation)
		}
	default:
		log.WithField("input-format", input.GetFormat()).Debug("unsupported input format")
		breach.EvaluateTemplate(p, &breach.ValueBreach{
			Value: fmt.Sprintf("unsupported input format %s", input.GetFormat()),
		}, nil)
	}
}
