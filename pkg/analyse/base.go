package analyse

import (
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	log "github.com/sirupsen/logrus"
)

// BaseAnalyser provides common fields and functionality for analyse plugins.
type BaseAnalyser struct {
	plugin.BasePlugin     `yaml:",inline"`
	Description           string `yaml:"description"`
	InputName             string `yaml:"input"`
	Severity              string `yaml:"severity"`
	breach.BreachTemplate `yaml:"breach-format"`
	Result                result.Result
	Remediation           interface{} `yaml:"remediation"`
	input                 fact.Facter
}

func (p *BaseAnalyser) GetDescription() string {
	return p.Description
}

func (p *BaseAnalyser) GetInputName() string {
	return p.InputName
}

func (p *BaseAnalyser) GetBreachTemplate() breach.BreachTemplate {
	return p.BreachTemplate
}

func (p *BaseAnalyser) GetResult() result.Result {
	if p.Description != "" && p.Result.Name != p.Description {
		p.Result.Name = p.Description
	}
	return p.Result
}

func (p *BaseAnalyser) SetInput(input fact.Facter) {
	p.input = input
}

func (p *BaseAnalyser) GetInput() fact.Facter {
	return p.input
}

func (p *BaseAnalyser) AddBreach(b breach.Breach) {
	b.SetCommonValues("", p.GetId(), p.Severity)
	p.Result.Breaches = append(p.Result.Breaches, b)
}

// Default implementations
func (p *BaseAnalyser) ValidateInput() error {
	log.WithFields(log.Fields{
		"analyser": p.Id,
	}).Debug("validating input")

	inPlugin := fact.Manager().FindPlugin(p.InputName)
	if inPlugin == nil {
		return &plugin.ErrSupportNotFound{
			Plugin: p.GetId(), SupportType: "input", SupportPlugin: p.InputName}
	}

	p.input = inPlugin
	return nil
}

func (p *BaseAnalyser) PreProcessInput() bool {
	if p.input == nil {
		p.AddBreach(&breach.ValueBreach{
			Value: "no input available to analyse",
		})
		return false
	}

	if len(p.input.GetErrors()) > 0 {
		errs := []string{}
		for _, e := range p.input.GetErrors() {
			errs = append(errs, e.Error())
		}
		p.AddBreach(&breach.KeyValuesBreach{
			Key:    "input failure",
			Values: errs,
		})
		return false
	}

	return true
}

func (p *BaseAnalyser) Analyse() {}
