package file

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

var (
	// ErrNoCurrentInput is returned when the plugin is configured without an
	// additional input naming the current (provisioned) file to compare
	// against the template.
	ErrNoCurrentInput = errors.New(
		"file:drift requires exactly one additional-inputs entry, naming the current file to compare against the template")
	// ErrTooManyAdditionalInputs is returned when more than one
	// additional-inputs entry is configured.
	ErrTooManyAdditionalInputs = errors.New(
		"file:drift accepts exactly one additional-inputs entry")
	// ErrInvalidPlaceholderPattern is returned when placeholder-pattern does
	// not compile as a regular expression.
	ErrInvalidPlaceholderPattern = errors.New("file:drift placeholder-pattern is not a valid regular expression")

	// defaultPlaceholderPattern matches Jinja/Mustache-style {{ VAR }}
	// placeholders, tolerating surrounding whitespace.
	defaultPlaceholderPattern = `{{\s*(\w+)\s*}}`
)

// Drift compares a template (its input) against the current state of a
// provisioned file (its single additional input), treating
// placeholder-shaped substrings (default `{{ VAR }}`) as wildcards rather
// than as literal text to diff. This is the general-purpose alternative to
// rendering the template with a vars map: it requires no per-project
// configuration, because it never needs to know what a placeholder's actual
// value should be - only that some value is present.
//
// Emits FormatListString: empty when the current file matches the template
// (modulo placeholders), or containing unified diff lines (plus any
// unsubstituted-placeholder notices) when it does not. Pairs naturally with
// the `drift` analyser, which breaches whenever this list is non-empty.
//
// Caveats (documented in docs/src/guide/gaps.md - Recipe: file:diff):
//   - Each placeholder is matched with a non-greedy (.*?) capture, so it can
//     absorb adjacent genuine drift on the same line rather than surfacing it.
//   - The same placeholder name repeated on one line is not checked for
//     consistency between its captures - RE2 has no backreferences.
//   - An empty substitution counts as substituted, not as drift; only a
//     placeholder still literally present in the current file is flagged.
type Drift struct {
	fact.BaseFact `yaml:",inline"`

	// Plugin fields.
	PlaceholderPattern string `yaml:"placeholder-pattern"`

	placeholderRegex *regexp.Regexp
}

func init() {
	fact.Manager().RegisterFactory("file:drift", func(n string) fact.Facter {
		return NewDrift(n)
	})
}

func NewDrift(id string) *Drift {
	return &Drift{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *Drift) GetName() string {
	return "file:drift"
}

// SupportedInputFormats declares that file:drift requires a template input
// of raw bytes or string (e.g. http:fetch or file:read).
func (p *Drift) SupportedInputFormats() (plugin.SupportLevel, []data.DataFormat) {
	return plugin.SupportRequired, []data.DataFormat{data.FormatRaw, data.FormatString}
}

func (p *Drift) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	additionalInputs := p.GetAdditionalInputs()
	if len(additionalInputs) == 0 {
		p.AddErrors(ErrNoCurrentInput)
		return
	}
	if len(additionalInputs) > 1 {
		p.AddErrors(ErrTooManyAdditionalInputs)
		return
	}
	currentInput := additionalInputs[0]

	// SupportedInputFormats() only constrains the primary input; the
	// additional (current-file) input is read directly via
	// stringifyFactData, which silently returns "" for any format it
	// doesn't recognise. Without this check, pointing additional-inputs at
	// e.g. a file:lookup (map-bytes) or json:key (list-string) output
	// produces an empty "current" side and a false-positive breach
	// claiming the entire template was deleted, rather than a config
	// error - the worst failure mode for an audit tool.
	switch currentInput.GetFormat() {
	case data.FormatRaw, data.FormatString:
	default:
		contextLogger.WithField("format", currentInput.GetFormat()).
			Error("unsupported additional input format")
		p.AddErrors(&plugin.ErrSupportNone{
			Plugin:        p.GetName(),
			SupportType:   "additionalInputFormat",
			SupportPlugin: string(currentInput.GetFormat())})
		return
	}

	pattern := p.PlaceholderPattern
	if pattern == "" {
		pattern = defaultPlaceholderPattern
	}
	placeholderRegex, err := regexp.Compile(pattern)
	if err != nil {
		contextLogger.WithError(err).Error("invalid placeholder pattern")
		p.AddErrors(ErrInvalidPlaceholderPattern)
		return
	}
	p.placeholderRegex = placeholderRegex

	templateContent := stringifyFactData(p.GetInput().GetData())
	currentContent := stringifyFactData(currentInput.GetData())

	drift := p.diff(templateContent, currentContent)

	p.Format = data.FormatListString
	p.SetData(drift)
}

// stringifyFactData accepts either raw bytes or string fact data, since
// file:drift's input may come from file:read (raw) or http:fetch (raw).
func stringifyFactData(d interface{}) string {
	switch v := d.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	default:
		return ""
	}
}

// diff masks placeholder-shaped substrings out of both the template and the
// current file before diffing, so legitimate per-project variation never
// appears as drift. A project-side placeholder that was never substituted is
// flagged as its own entry, since forward-rendering approaches cannot detect
// this common provisioning bug.
func (p *Drift) diff(templateContent, currentContent string) []string {
	templateLines := difflib.SplitLines(templateContent)
	currentLines := difflib.SplitLines(currentContent)

	maskedTemplate := make([]string, len(templateLines))
	maskedCurrent := make([]string, len(currentLines))

	matcher := difflib.NewMatcher(templateLines, currentLines)
	var unsubstituted []string

	for _, op := range matcher.GetOpCodes() {
		switch op.Tag {
		case 'e':
			for k := 0; k < op.I2-op.I1; k++ {
				tLine := templateLines[op.I1+k]
				maskedTemplate[op.I1+k] = tLine
				maskedCurrent[op.J1+k] = tLine
				// An equal line containing a placeholder means the current
				// file still literally contains the placeholder text - it
				// was never substituted. Rendering-based approaches cannot
				// detect this, because there is nothing on the current side
				// to compare against once the placeholder is rendered away.
				for _, ph := range p.placeholderRegex.FindAllString(tLine, -1) {
					unsubstituted = append(unsubstituted, fmt.Sprintf(
						"line %d: placeholder %q was left unsubstituted", op.J1+k+1, ph))
				}
			}
		case 'r':
			n := op.I2 - op.I1
			if n == op.J2-op.J1 {
				for k := 0; k < n; k++ {
					tLine := templateLines[op.I1+k]
					cLine := currentLines[op.J1+k]
					masked, values, matched := maskLine(p.placeholderRegex, tLine, cLine)
					maskedTemplate[op.I1+k] = tLine
					if matched {
						maskedCurrent[op.J1+k] = masked
						for _, v := range values {
							if p.placeholderRegex.MatchString(v) {
								unsubstituted = append(unsubstituted, fmt.Sprintf(
									"line %d: placeholder %q was left unsubstituted", op.J1+k+1, v))
							}
						}
					} else {
						maskedCurrent[op.J1+k] = cLine
					}
				}
			} else {
				for k := 0; k < n; k++ {
					maskedTemplate[op.I1+k] = templateLines[op.I1+k]
				}
				for k := 0; k < op.J2-op.J1; k++ {
					maskedCurrent[op.J1+k] = currentLines[op.J1+k]
				}
			}
		case 'd':
			for k := 0; k < op.I2-op.I1; k++ {
				maskedTemplate[op.I1+k] = templateLines[op.I1+k]
			}
		case 'i':
			for k := 0; k < op.J2-op.J1; k++ {
				maskedCurrent[op.J1+k] = currentLines[op.J1+k]
			}
		}
	}

	unifiedDiff := difflib.UnifiedDiff{
		A:        maskedTemplate,
		B:        maskedCurrent,
		FromFile: "template",
		ToFile:   "current",
		Context:  3,
	}
	diffStr, _ := difflib.GetUnifiedDiffString(unifiedDiff)

	var result []string
	if diffStr != "" {
		// SplitLines preserves each line's trailing "\n", which is what
		// UnifiedDiff needs as input elsewhere - but here each entry becomes
		// its own list item, and the "drift" analyser's breach joins items
		// with its own separator, so a retained "\n" would double up into a
		// blank line between every entry. SplitLines also always appends an
		// artificial trailing "\n"-only entry when the input ends in a
		// newline (its convention for reconstructing multi-line text), which
		// becomes an empty string once trimmed - drop it rather than emit a
		// dangling bullet in rendered breach output.
		for _, line := range difflib.SplitLines(diffStr) {
			trimmed := strings.TrimSuffix(line, "\n")
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	result = append(result, unsubstituted...)
	return result
}

// maskLine attempts to match currentLine against templateLine, treating each
// placeholder in templateLine as a wildcard capture. On a successful match,
// it returns templateLine (so the diff sees no change on this line) and the
// captured values (so callers can detect unsubstituted placeholders).
func maskLine(placeholderRegex *regexp.Regexp, templateLine, currentLine string) (masked string, values []string, matched bool) {
	locs := placeholderRegex.FindAllStringIndex(templateLine, -1)
	if len(locs) == 0 {
		return templateLine, nil, templateLine == currentLine
	}

	linePattern := "^"
	lastEnd := 0
	for _, loc := range locs {
		linePattern += regexp.QuoteMeta(templateLine[lastEnd:loc[0]])
		linePattern += "(.*?)"
		lastEnd = loc[1]
	}
	linePattern += regexp.QuoteMeta(templateLine[lastEnd:]) + "$"

	lineRegex, err := regexp.Compile(linePattern)
	if err != nil {
		return templateLine, nil, false
	}

	m := lineRegex.FindStringSubmatch(currentLine)
	if m == nil {
		return templateLine, nil, false
	}
	return templateLine, m[1:], true
}
