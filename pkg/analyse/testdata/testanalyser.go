package testdata

import (
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type TestAnalyser struct {
	// Common fields.
	Id                    string `yaml:"name"`
	Description           string `yaml:"description"`
	InputName             string `yaml:"input"`
	breach.BreachTemplate `yaml:"breach-format"`
	Result                result.Result
}

// Common plugin methods.
func (p *TestAnalyser) PluginName() string { return "test-analyser" }

func (p *TestAnalyser) GetId() string { return p.Id }

// Analyse methods.

func (p *TestAnalyser) GetDescription() string { return p.Description }

func (p *TestAnalyser) GetInputName() string { return p.InputName }

func (p *TestAnalyser) GetBreachTemplate() breach.BreachTemplate { return p.BreachTemplate }

func (p *TestAnalyser) GetResult() result.Result { return p.Result }

func (p *TestAnalyser) ValidateInput() error { return nil }

func (p *TestAnalyser) PreProcessInput() bool { return true }

func (p *TestAnalyser) Analyse() {}

func (p *TestAnalyser) AddBreach(b breach.Breach) {}
