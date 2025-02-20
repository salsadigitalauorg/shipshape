package testdata

import "github.com/salsadigitalauorg/shipshape/pkg/analyse"

type TestAnalyser struct {
	analyse.BaseAnalyser
}

func (p *TestAnalyser) GetName() string       { return "test-analyser" }
func (p *TestAnalyser) ValidateInput() error  { return nil }
func (p *TestAnalyser) PreProcessInput() bool { return true }
