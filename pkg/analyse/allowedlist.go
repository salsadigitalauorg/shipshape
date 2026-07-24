package analyse

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

type AllowedList struct {
	BaseAnalyser `yaml:",inline"`
	PackageMatch string `yaml:"package-match"`
	pkgRegex     *regexp.Regexp
	Allowed      []string `yaml:"allowed"`
	Required     []string `yaml:"required"`
	Deprecated   []string `yaml:"deprecated"`
	ExcludeKeys  []string `yaml:"exclude-keys"`
	Ignore       []string `yaml:"ignore"`

	// If a map is provided as input, Key is used to look up the value.
	Key string `yaml:"key"`

	// If NotStrict is false, only values that are in the Allowed list are allowed.
	NotStrict bool `yaml:"not-strict"`
}

//go:generate go run ../../cmd/gen.go analyse-plugin --plugin=AllowedList --package=analyse

func init() {
	Manager().RegisterFactory("allowed:list", func(id string) Analyser {
		return NewAllowedList(id)
	})
}

func (p *AllowedList) GetName() string {
	return "allowed:list"
}

func (p *AllowedList) Analyse() {
	if p.PackageMatch != "" {
		p.pkgRegex = regexp.MustCompile("^(.[^:@]*)?[:@]?([^ latest$]*)")
	}

	switch p.input.GetFormat() {
	case data.FormatListString:
		inputData := data.AsListString(p.input.GetData())
		foundRequired := map[string]bool{}
		for _, v := range inputData {
			if p.isIgnored(v) {
				continue
			}

			if p.isDeprecated(v) {
				breach.EvaluateTemplate(p, &breach.ValueBreach{
					ValueLabel: "deprecated value found",
					Value:      v,
				}, p.Remediation)
				continue
			}

			if !p.NotStrict && len(p.Required) == 0 && !p.isAllowed(v) {
				breach.EvaluateTemplate(p, &breach.ValueBreach{
					ValueLabel: "disallowed value found",
					Value:      v,
				}, p.Remediation)
			}

			if len(p.Required) > 0 && p.isRequired(v) {
				foundRequired[v] = true
			}
		}

		if len(p.Required) > 0 {
			for _, r := range p.Required {
				if ok := foundRequired[r]; !ok {
					breach.EvaluateTemplate(p, &breach.ValueBreach{
						ValueLabel: "required value not found",
						Value:      r,
					}, p.Remediation)
				}
			}
		}
	case data.FormatMapString:
		inputData := data.AsMapString(p.input.GetData())
		foundRequired := map[string]bool{}

		// If Key is specified, only process that specific key-value pair
		if p.Key != "" {
			v, exists := inputData[p.Key]
			if !exists {
				log.WithField("key", p.Key).Warning("specified key not found in input data")
				return
			}

			// Split multiline string into individual lines and process each line
			lines := strings.Split(v, "\n")
			for _, line := range lines {
				// Trim whitespace from each line
				line = strings.TrimSpace(line)

				// Skip empty lines
				if line == "" {
					continue
				}

				if p.isIgnored(line) {
					continue
				}

				if p.isDeprecated(line) {
					breach.EvaluateTemplate(p, &breach.ValueBreach{
						ValueLabel: "deprecated value found",
						Value:      line,
					}, p.Remediation)
					continue
				}

				if !p.NotStrict && len(p.Required) == 0 && !p.isAllowed(line) {
					breach.EvaluateTemplate(p, &breach.ValueBreach{
						ValueLabel: "disallowed value found",
						Value:      line,
					}, p.Remediation)
				}

				if len(p.Required) > 0 && p.isRequired(line) {
					foundRequired[line] = true
				}
			}
		} else {
			// If Key is not specified, process all key-value pairs (original behavior)
			for k, v := range inputData {
				if p.isExcludedKey(k) || p.isIgnored(v) {
					continue
				}

				if p.isDeprecated(v) {
					breach.EvaluateTemplate(p, &breach.KeyValueBreach{
						KeyLabel:   "key",
						Key:        k,
						ValueLabel: "deprecated",
						Value:      v,
					}, p.Remediation)
					continue
				}

				if !p.NotStrict && len(p.Required) == 0 && !p.isAllowed(v) {
					breach.EvaluateTemplate(p, &breach.KeyValueBreach{
						KeyLabel:   "key",
						Key:        k,
						ValueLabel: "disallowed",
						Value:      v,
					}, p.Remediation)
					continue
				}

				if len(p.Required) > 0 && p.isRequired(v) {
					foundRequired[v] = true
				}
			}
		}

		if len(p.Required) > 0 {
			for _, r := range p.Required {
				if ok := foundRequired[r]; !ok {
					breach.EvaluateTemplate(p, &breach.ValueBreach{
						ValueLabel: "required value not found",
						Value:      r,
					}, p.Remediation)
				}
			}
		}

	case data.FormatMapListString:
		inputData := data.AsMapListString(p.input.GetData())
		for k, listV := range inputData {
			if p.isExcludedKey(k) {
				continue
			}

			foundRequired := map[string]bool{}
			for _, v := range listV {
				if p.isIgnored(v) {
					continue
				}

				if p.isDeprecated(v) {
					breach.EvaluateTemplate(p, &breach.KeyValueBreach{
						KeyLabel:   "key",
						Key:        k,
						ValueLabel: "deprecated",
						Value:      v,
					}, p.Remediation)
					continue
				}

				if len(p.Required) == 0 && !p.isAllowed(v) {
					breach.EvaluateTemplate(p, &breach.KeyValueBreach{
						KeyLabel:   "key",
						Key:        k,
						ValueLabel: "disallowed",
						Value:      v,
					}, p.Remediation)
					continue
				}

				if len(p.Required) > 0 && p.isRequired(v) {
					foundRequired[v] = true
				}
			}

			if len(p.Required) > 0 {
				for _, r := range p.Required {
					if ok := foundRequired[r]; !ok {
						breach.EvaluateTemplate(p, &breach.KeyValueBreach{
							Key:        k,
							ValueLabel: "required value not found",
							Value:      r,
						}, p.Remediation)
					}
				}
			}
		}

	default:
		log.WithField("input-format", p.input.GetFormat()).Error("unsupported input format")
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

func (p *AllowedList) isRequired(value string) bool {
	for _, i := range p.Required {
		if i == value {
			return true
		}
	}
	return false
}
