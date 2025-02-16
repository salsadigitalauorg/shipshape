package analyse

import (
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type Analyser interface {
	plugin.Plugin

	// Input methods
	SetInput(input fact.Facter)
	GetInput() fact.Facter
	GetInputName() string
	ValidateInput() error
	PreProcessInput() bool

	// Analysis methods
	GetDescription() string
	GetBreachTemplate() breach.BreachTemplate
	GetResult() result.Result
	Analyse()
	AddBreach(b breach.Breach)
}
