package yaml

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

// typeErrorPrefix labels a *yaml.TypeError, distinguishing a document that
// parsed but contained conflicting/duplicate entries from an outright syntax
// error. It mirrors the two breach shapes the 0.x yamllint check reported
// (pkg/checks/yaml/yamllintcheck.go), which are preserved here so migrated
// configs surface a recognisable message.
const typeErrorPrefix = "cannot decode yaml: "

// Lint reports which of its input files fail to parse as YAML.
//
// It inverts the usual fact/error relationship: everywhere else in the
// pipeline a YAML parse failure is a *collection error*, and any collection
// error is fatal to the run (pkg/shipshape/shipshape.go), so the pipeline
// never reaches an analyser. That makes the exact condition the 0.x yamllint
// check exists to report unreachable in 1.x. yaml:lint therefore treats a
// parse failure as ordinary *data* - a file->message map - so an analyser can
// act on it. That inversion is the whole point of the plugin, and it is why
// Collect never calls AddErrors for a malformed document; the only errors it
// raises are genuine misconfiguration (an unusable input format), which
// should stay fatal.
//
// Emits FormatMapString containing only the failures, so an entirely valid
// tree emits an empty map. Pairs with the not:empty analyser, which breaches
// once per entry.
type Lint struct {
	fact.BaseFact `yaml:",inline"`

	// AllDocuments validates every document in a multi-document
	// ('---'-separated) file rather than only the first. Defaults to true,
	// which is a deliberate improvement on 0.x: that check used a single
	// yaml.Unmarshal, which stops after the first document and so silently
	// passed a file whose later documents were malformed. Set false for
	// bug-for-bug 0.x parity.
	//
	// A pointer so an operator explicitly setting 'false' is
	// distinguishable from the field being absent, matching how
	// IgnoreMissing is handled on the 0.x yaml checks.
	AllDocuments *bool `yaml:"all-documents"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --package=yaml

func init() {
	fact.Manager().RegisterFactory("yaml:lint", func(n string) fact.Facter {
		return NewLint(n)
	})
}

func NewLint(id string) *Lint {
	allDocuments := true
	return &Lint{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
		AllDocuments: &allDocuments,
	}
}

func (p *Lint) GetName() string {
	return "yaml:lint"
}

// SupportedInputFormats accepts only FormatMapBytes, the format emitted by
// file:lookup with 'file-names-only: false'. Delegating file selection to
// file:lookup means yaml:lint inherits path/pattern/exclude-pattern/skip-dirs
// rather than reimplementing them.
func (p *Lint) SupportedInputFormats() (plugin.SupportLevel, []data.DataFormat) {
	return plugin.SupportRequired, []data.DataFormat{data.FormatMapBytes}
}

// allDocuments resolves the AllDocuments setting, defaulting to true when the
// field is absent. A Lint built by yaml.Unmarshal rather than NewLint (as the
// fact manager does when a config omits the key) leaves the pointer nil.
func (p *Lint) allDocuments() bool {
	return p.AllDocuments == nil || *p.AllDocuments
}

func (p *Lint) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	contextLogger.WithFields(log.Fields{
		"input":         p.GetInputName(),
		"input-plugin":  p.GetInput().GetName(),
		"input-format":  p.GetInput().GetFormat(),
		"all-documents": p.allDocuments(),
	}).Debug("collecting data")

	switch p.GetInput().GetFormat() {
	case data.FormatMapBytes:
		inputData := data.AsMapBytes(p.GetInput().GetData())

		// Always emit map-string, even for no input or an all-valid tree, so
		// not:empty sees a supported (empty) format rather than logging an
		// unsupported-format error.
		invalid := map[string]string{}
		for f, src := range inputData {
			if msg, ok := p.validate(src); !ok {
				invalid[f] = msg
			}
		}

		contextLogger.WithFields(log.Fields{
			"files-checked": len(inputData),
			"files-invalid": len(invalid),
		}).Debug("linted yaml")

		p.Format = data.FormatMapString
		p.SetData(invalid)

	default:
		contextLogger.WithField("format", p.GetInput().GetFormat()).
			Error("unsupported input format")
		p.AddErrors(fmt.Errorf("unsupported input format %s", p.GetInput().GetFormat()))
	}
}

// validate reports whether src parses as YAML, returning a formatted message
// if not. Decoding into interface{} keeps this a pure syntax gate, making no
// assertion about the document's contents - that is yaml:key's job.
//
// Returns ok=false with the message when src is invalid.
//
// Note on line numbers: the message comes verbatim from gopkg.in/yaml.v3,
// which for an unterminated construct reports the line where that construct
// *opened* rather than where parsing gave up. An unclosed '[' on line 4 of a
// 6-line file is reported as line 3. This is the library's behaviour and is
// identical between Unmarshal and Decoder (verified), so it is passed through
// unaltered rather than "corrected" - second-guessing the parser's own
// position would be guesswork, and the message is still enough to locate the
// fault. The 0.x check passed the same messages through.
func (p *Lint) validate(src []byte) (string, bool) {
	if !p.allDocuments() {
		var ifc interface{}
		if err := yaml.Unmarshal(src, &ifc); err != nil {
			return formatErr(err), false
		}
		return "", true
	}

	// A decoder loop reaches every document in a '---'-separated file.
	// yaml.Unmarshal would return after the first, so a malformed later
	// document would go unreported.
	dec := yaml.NewDecoder(bytes.NewReader(src))
	for {
		var ifc interface{}
		err := dec.Decode(&ifc)
		if errors.Is(err, io.EOF) {
			return "", true
		}
		if err != nil {
			return formatErr(err), false
		}
	}
}

// formatErr renders a parse error as a single-line message, preserving the
// 0.x distinction between a type error and any other parse failure. A
// *yaml.TypeError aggregates several messages, which are joined rather than
// dropped so every conflict in the document is reported.
func formatErr(err error) string {
	var typeErr *yaml.TypeError
	if errors.As(err, &typeErr) {
		return typeErrorPrefix + strings.Join(typeErr.Errors, "; ")
	}
	return err.Error()
}
