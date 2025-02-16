package testdata

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/analyse"
)

type TestAnalyserInputError struct {
	analyse.BaseAnalyser
}

func (p *TestAnalyserInputError) GetName() string       { return "test-analyser" }
func (p *TestAnalyserInputError) ValidateInput() error  { return errors.New("input error") }
func (p *TestAnalyserInputError) PreProcessInput() bool { return true }
