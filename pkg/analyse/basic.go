package analyse

import (
	"github.com/salsadigitalauorg/shipshape/pkg/conditions"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type Basic struct {
	// Common fields.
	Name      string `yaml:"name"`
	InputName string `yaml:"input"`
	Severity  string `yaml:"severity"`
	Result    result.Result
	input     fact.Facter

	// Plugin fields.
	Empty              bool   `yaml:"empty"`
	NotEmpty           bool   `yaml:"not-empty"`
	Equal              string `yaml:"equal"`
	NotEqual           string `yaml:"not-equal"`
	RegexMatch         string `yaml:"regex-match"`
	Skip               string `yaml:"skip"`
	conditionsResolver conditions.Resolver
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=Basic --package=analyse

func init() {
	Registry["basic"] = func(n string) Analyser { return NewBasic(n) }
}

func (p *Basic) PluginName() string {
	return "basic"
}

func (p *Basic) BuildConditionsResolver() {
	p.conditionsResolver = conditions.Resolver{}
	if p.Empty {
		p.conditionsResolver.AddCondition(&conditions.Empty{})
	}
	if p.NotEmpty {
		p.conditionsResolver.AddCondition(&conditions.NotEmpty{})
	}
	if p.Equal != "" {
		p.conditionsResolver.AddCondition(&conditions.Equal{})
	}
	if p.NotEqual != "" {
		p.conditionsResolver.AddCondition(&conditions.NotEqual{})
	}
	if p.RegexMatch != "" {
		p.conditionsResolver.AddCondition(&conditions.RegexMatch{
			Regex: p.RegexMatch, Skip: p.Skip})
	}
}

func (p *Basic) Analyse() {
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
				_, err := p.conditionsResolver.Resolve(v)
				if err != nil {
					p.AddBreach(&result.KeyValueBreach{
						Key:        k,
						ValueLabel: k2,
						Value:      err.Error(),
					})
					continue
				}
			}
		}
	}
}
