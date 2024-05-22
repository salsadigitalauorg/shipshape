package testdata

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type TestAnalyserPreprocessInputFail struct {
	// Common fields.
	Id                    string `yaml:"name"`
	Description           string `yaml:"description"`
	InputName             string `yaml:"input"`
	Severity              string `yaml:"severity"`
	breach.BreachTemplate `yaml:"breach-format"`
	Result                result.Result
}

func (p *TestAnalyserPreprocessInputFail) PluginName() string { return "test-analyser" }

func (p *TestAnalyserPreprocessInputFail) GetId() string { return p.Id }

func (p *TestAnalyserPreprocessInputFail) GetDescription() string { return p.Description }

func (p *TestAnalyserPreprocessInputFail) GetInputName() string { return p.InputName }

func (p *TestAnalyserPreprocessInputFail) GetBreachTemplate() breach.BreachTemplate {
	return p.BreachTemplate
}

func (p *TestAnalyserPreprocessInputFail) GetResult() result.Result { return p.Result }

func (p *TestAnalyserPreprocessInputFail) ValidateInput() error { return errors.New("input error") }

func (p *TestAnalyserPreprocessInputFail) PreProcessInput() bool {
	p.AddBreach(&breach.KeyValuesBreach{
		Key:    "input failure",
		Values: []string{"input error"},
	})
	return false
}

func (p *TestAnalyserPreprocessInputFail) Analyse() {}

func (p *TestAnalyserPreprocessInputFail) AddBreach(b breach.Breach) {
	b.SetCommonValues("", p.Id, p.Severity)
	p.Result.Breaches = append(
		p.Result.Breaches,
		b,
	)
}
