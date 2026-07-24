package json

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/PaesslerAG/jsonpath"
	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

var (
	// ErrNoExpression is returned when the plugin is configured without a
	// JSONPath expression.
	ErrNoExpression = errors.New("json:key requires an 'expression'")
	// ErrInvalidJSON is returned when the input cannot be decoded as JSON.
	ErrInvalidJSON = errors.New("invalid JSON input")
)

// Key evaluates a JSONPath expression against JSON input and emits the matched
// values. It is the JSON counterpart to yaml:key, using the JSONPath dialect
// (e.g. "$.a.b[0]", "$.items.*") rather than a dotted-path lookup.
type Key struct {
	fact.BaseFact `yaml:",inline"`

	// Expression is the JSONPath expression to evaluate against the input.
	Expression string `yaml:"expression"`
}

func init() {
	fact.Manager().RegisterFactory("json:key", func(n string) fact.Facter {
		return New(n)
	})
}

func New(id string) *Key {
	return &Key{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *Key) GetName() string {
	return "json:key"
}

// SupportedInputFormats declares that json:key requires an input and reads
// either raw JSON bytes (file:read) or a map of JSON bytes keyed by filename
// (file:lookup).
func (p *Key) SupportedInputFormats() (plugin.SupportLevel, []data.DataFormat) {
	return plugin.SupportRequired, []data.DataFormat{
		data.FormatRaw,
		data.FormatMapBytes,
	}
}

func (p *Key) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	if p.Expression == "" {
		p.AddErrors(ErrNoExpression)
		return
	}

	contextLogger.WithFields(log.Fields{
		"input":        p.GetInputName(),
		"input-plugin": p.GetInput().GetName(),
		"input-format": p.GetInput().GetFormat(),
		"expression":   p.Expression,
	}).Debug("collecting data")

	switch p.GetInput().GetFormat() {
	case data.FormatRaw:
		inputData := data.AsBytes(p.GetInput().GetData())
		if inputData == nil {
			return
		}

		values, err := p.query(inputData)
		if err != nil {
			contextLogger.WithError(err).Error("error evaluating jsonpath")
			p.AddErrors(sentinel(err))
			return
		}

		p.Format = data.FormatListString
		p.SetData(values)

	case data.FormatMapBytes:
		inputData := data.AsMapBytes(p.GetInput().GetData())
		if inputData == nil {
			return
		}

		result := map[string][]string{}
		for name, b := range inputData {
			values, err := p.query(b)
			if err != nil {
				contextLogger.WithError(err).WithField("file", name).
					Error("error evaluating jsonpath")
				p.AddErrors(sentinel(err))
				return
			}
			result[name] = values
		}

		p.Format = data.FormatMapListString
		p.SetData(result)

	default:
		contextLogger.WithField("format", p.GetInput().GetFormat()).
			Error("unsupported input format")
		p.AddErrors(fmt.Errorf("unsupported input format %s", p.GetInput().GetFormat()))
	}
}

// sentinel maps a detailed internal error back to a stable, comparable
// sentinel so downstream reporting is deterministic. Errors that do not match a
// known sentinel are returned unchanged.
func sentinel(err error) error {
	if errors.Is(err, ErrInvalidJSON) {
		return ErrInvalidJSON
	}
	return err
}

// query decodes JSON bytes, evaluates the JSONPath expression, and normalises
// the result to a sorted slice of strings. Sorting keeps the output
// deterministic even when the expression matches an (unordered) JSON object's
// values via a wildcard.
func (p *Key) query(src []byte) ([]string, error) {
	var doc interface{}
	if err := stdjson.Unmarshal(src, &doc); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidJSON, err)
	}

	match, err := jsonpath.Get(p.Expression, doc)
	if err != nil {
		return nil, fmt.Errorf("jsonpath %q: %w", p.Expression, err)
	}

	values := flatten(match)
	sort.Strings(values)
	return values, nil
}

// flatten converts a JSONPath match into a slice of strings. A wildcard or
// filter match yields a []interface{}; a single-node match yields a scalar.
func flatten(match interface{}) []string {
	switch v := match.(type) {
	case nil:
		return []string{}
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, scalarToString(item))
		}
		return out
	default:
		return []string{scalarToString(v)}
	}
}

// scalarToString renders a JSON scalar as a string. Numbers decoded by
// encoding/json are float64; render integral values without a trailing ".0".
func scalarToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case bool:
		return fmt.Sprintf("%t", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
