package analyse

import (
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

type Analyser interface {
	// Common plugin methods.
	PluginName() string
	GetId() string

	// Analyse methods.
	GetDescription() string
	GetInputName() string
	GetBreachTemplate() breach.BreachTemplate
	GetResult() result.Result
	ValidateInput() error
	PreProcessInput() bool
	Analyse()
	AddBreach(b breach.Breach)
}
