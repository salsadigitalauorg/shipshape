package analyse

import "github.com/salsadigitalauorg/shipshape/pkg/result"

type Analyser interface {
	// Common plugin methods.
	PluginName() string
	GetName() string

	// Analyse methods.
	GetInputName() string
	GetResult() result.Result
	ValidateInput() error
	Analyse()
	AddBreach(b result.Breach)
}
