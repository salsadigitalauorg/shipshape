package testdata

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
)

type TestAnalyserPreprocessInputFail struct {
	analyse.BaseAnalyser
}

func (p *TestAnalyserPreprocessInputFail) GetName() string { return "test-analyser" }

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
