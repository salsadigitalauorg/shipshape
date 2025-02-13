package analyse

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type RegexMatch struct {
	// Common fields.
	Id                    string `yaml:"name"`
	Description           string `yaml:"description"`
	InputName             string `yaml:"input"`
	Severity              string `yaml:"severity"`
	breach.BreachTemplate `yaml:"breach-format"`
	Result                result.Result
	Remediation           interface{} `yaml:"remediation"`
	input                 fact.Facter

	// Plugin fields.
	Pattern string `yaml:"pattern"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=RegexMatch --package=analyse

func init() {
	Registry["regex:match"] = func(id string) Analyser { return NewRegexMatch(id) }
}

func (p *RegexMatch) PluginName() string {
	return "regex:match"
}

func (p *RegexMatch) Analyse() {
	switch p.input.GetFormat() {
	case data.FormatNil:
		return
	case data.FormatMapNestedString:
		inputData := data.AsMapNestedString(p.input.GetData())
		for k, kvs := range inputData {
			for k2, v := range kvs {
				match, _ := regexp.MatchString(p.Pattern, v)
				if match {
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
		if match {
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
