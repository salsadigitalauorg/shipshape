package testdata

import (
	"github.com/salsadigitalauorg/shipshape/pkg/analyse"
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
)

type TestAnalyserPass struct {
	analyse.BaseAnalyser
}

func (p *TestAnalyserPass) GetName() string { return "test-analyser" }

func (p *TestAnalyserPass) ValidateInput() error { return nil }

func (p *TestAnalyserPass) PreProcessInput() bool { return true }

func (p *TestAnalyserPass) Analyse() {
	p.AddBreach(&breach.KeyValuesBreach{
		Key:    "breach found",
		Values: []string{"more details would be here"},
	})
}
