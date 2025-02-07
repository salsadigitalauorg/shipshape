package testdata

import (
	"errors"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type TestAnalyserInputError struct {
	// Common fields.
	Id                    string `yaml:"name"`
	Description           string `yaml:"description"`
	InputName             string `yaml:"input"`
	breach.BreachTemplate `yaml:"breach-format"`
	Result                result.Result
	input                 fact.Facter
}

// Common plugin methods.
func (p *TestAnalyserInputError) PluginName() string { return "test-analyser" }

func (p *TestAnalyserInputError) GetId() string { return p.Id }

// Analyse methods.

func (p *TestAnalyserInputError) SetInput(input fact.Facter) { p.input = input }

func (p *TestAnalyserInputError) GetDescription() string { return p.Description }

func (p *TestAnalyserInputError) GetInputName() string { return p.InputName }

func (p *TestAnalyserInputError) GetBreachTemplate() breach.BreachTemplate { return p.BreachTemplate }

func (p *TestAnalyserInputError) GetResult() result.Result { return p.Result }

func (p *TestAnalyserInputError) ValidateInput() error { return errors.New("input error") }

func (p *TestAnalyserInputError) PreProcessInput() bool { return true }

func (p *TestAnalyserInputError) Analyse() {}

func (p *TestAnalyserInputError) AddBreach(b breach.Breach) {}
