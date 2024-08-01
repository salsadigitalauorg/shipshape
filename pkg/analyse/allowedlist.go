package analyse

import (
	"regexp"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type AllowedList struct {
	// Common fields.
	Id                    string `yaml:"name"`
	Description           string `yaml:"description"`
	InputName             string `yaml:"input"`
	Severity              string `yaml:"severity"`
	breach.BreachTemplate `yaml:"breach-format"`
	Result                result.Result
	input                 fact.Facter

	// Plugin fields.
	PackageMatch string `yaml:"package-match"`
	pkgRegex     *regexp.Regexp
	Allowed      []string `yaml:"allowed"`
	Deprecated   []string `yaml:"deprecated"`
	ExcludeKeys  []string `yaml:"exclude-keys"`
	Ignore       []string `yaml:"ignore"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=AllowedList --package=analyse

func init() {
	Registry["allowed:list"] = func(id string) Analyser { return NewAllowedList(id) }
}

func (p *AllowedList) PluginName() string {
	return "allowed:list"
}

func (p *AllowedList) Analyse() {
	if p.PackageMatch != "" {
		p.pkgRegex = regexp.MustCompile("^(.[^:@]*)?[:@]?([^ latest$]*)")
	}

	switch p.input.GetFormat() {
	case data.FormatMapString:
		inputData := data.AsMapString(p.input.GetData())
		for k, v := range inputData {
			if p.isExcludedKey(k) || p.isIgnored(v) {
				continue
			}

			if !p.isAllowed(v) {
				breach.EvaluateTemplate(p, &breach.KeyValueBreach{
					KeyLabel:   "key",
					Key:        k,
					ValueLabel: "disallowed",
					Value:      v,
				})
				continue
			}

			if p.isDeprecated(v) {
				breach.EvaluateTemplate(p, &breach.KeyValueBreach{
					KeyLabel:   "key",
					Key:        k,
					ValueLabel: "deprecated",
					Value:      v,
				})
				continue
			}
		}
	case data.FormatMapListString:
		inputData := data.AsMapListString(p.input.GetData())
		for k, listV := range inputData {
			if p.isExcludedKey(k) || p.isIgnored(k) {
				continue
			}

			for _, v := range listV {
				if !p.isAllowed(v) {
					breach.EvaluateTemplate(p, &breach.KeyValueBreach{
						KeyLabel:   "key",
						Key:        k,
						ValueLabel: "disallowed",
						Value:      v,
					})
					continue
				}

				if p.isDeprecated(v) {
					breach.EvaluateTemplate(p, &breach.KeyValueBreach{
						KeyLabel:   "key",
						Key:        k,
						ValueLabel: "deprecated",
						Value:      v,
					})
					continue
				}
			}
		}
	}
}

func (p *AllowedList) isAllowed(value string) bool {
	if p.pkgRegex != nil {
		match := p.pkgRegex.FindStringSubmatch(value)
		if len(match) < 1 {
			return false
		}

		if !utils.PackageCheckString(p.Allowed, match[1], match[2]) {
			return false
		}
		return true
	}
	for _, a := range p.Allowed {
		if a == value {
			return true
		}
	}
	return false
}

func (p *AllowedList) isDeprecated(value string) bool {
	for _, d := range p.Deprecated {
		if d == value {
			return true
		}
	}
	return false
}

func (p *AllowedList) isExcludedKey(key string) bool {
	for _, e := range p.ExcludeKeys {
		if e == key {
			return true
		}
	}
	return false
}

func (p *AllowedList) isIgnored(value string) bool {
	for _, i := range p.Ignore {
		if i == value {
			return true
		}
	}
	return false
}
