package testdata

import (
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type TestAnalyserPass struct {
	// Common fields.
	Id                    string `yaml:"name"`
	Description           string `yaml:"description"`
	InputName             string `yaml:"input"`
	Severity              string `yaml:"severity"`
	breach.BreachTemplate `yaml:"breach-format"`
	Result                result.Result
}

func (p *TestAnalyserPass) PluginName() string { return "test-analyser" }

func (p *TestAnalyserPass) GetId() string { return p.Id }

func (p *TestAnalyserPass) GetDescription() string { return p.Description }

func (p *TestAnalyserPass) GetInputName() string { return p.InputName }

func (p *TestAnalyserPass) GetBreachTemplate() breach.BreachTemplate {
	return p.BreachTemplate
}

func (p *TestAnalyserPass) GetResult() result.Result { return p.Result }

func (p *TestAnalyserPass) ValidateInput() error { return nil }

func (p *TestAnalyserPass) PreProcessInput() bool { return true }

func (p *TestAnalyserPass) Analyse() {
	p.AddBreach(&breach.KeyValuesBreach{
		Key:    "breach found",
		Values: []string{"more details would be here"},
	})
}

func (p *TestAnalyserPass) AddBreach(b breach.Breach) {
	b.SetCommonValues("", p.Id, p.Severity)
	p.Result.Breaches = append(
		p.Result.Breaches,
		b,
	)
}
