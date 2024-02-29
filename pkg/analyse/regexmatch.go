package analyse

import (
	"regexp"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type RegexMatch struct {
	// Common fields.
	Name      string `yaml:"name"`
	InputName string `yaml:"input"`
	Severity  string `yaml:"severity"`
	Result    result.Result
	input     fact.Facter

	// Plugin fields.
	Pattern string `yaml:"pattern"`
	Ignore  string `yaml:"ignore"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=RegexMatch --package=analyse

func init() {
	Registry["regex-match"] = func(n string) Analyser { return NewRegexMatch(n) }
}

func (p *RegexMatch) PluginName() string {
	return "regex-match"
}

func (p *RegexMatch) Analyse() {
	if p.input == nil {
		p.AddBreach(&result.ValueBreach{
			Value: "no input available to analyse",
		})
		return
	}

	switch p.input.GetFormat() {
	case fact.FormatMapNestedString:
		inputData := data.AsNestedStringMap(p.input.GetData())
		for k, kvs := range inputData {
			for k2, v := range kvs {
				match, _ := regexp.MatchString(p.Pattern, v)
				if match {
					p.AddBreach(&result.KeyValueBreach{
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
