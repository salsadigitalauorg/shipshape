package analyse

import (
	"regexp"

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
	input                 fact.Facter

	// Plugin fields.
	Pattern string `yaml:"pattern"`
	Ignore  string `yaml:"ignore"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=RegexMatch --package=analyse

func init() {
	Registry["regex-match"] = func(id string) Analyser { return NewRegexMatch(id) }
}

func (p *RegexMatch) PluginName() string {
	return "regex-match"
}

func (p *RegexMatch) Analyse() {
	switch p.input.GetFormat() {
	case data.FormatMapNestedString:
		inputData := data.AsNestedStringMap(p.input.GetData())
		for k, kvs := range inputData {
			for k2, v := range kvs {
				match, _ := regexp.MatchString(p.Pattern, v)
				if match {
					breach.EvaluateTemplate(p, &breach.KeyValueBreach{
						Key:        k,
						ValueLabel: k2,
						Value:      v,
					})
					continue
				}
			}
		}
	}
}
