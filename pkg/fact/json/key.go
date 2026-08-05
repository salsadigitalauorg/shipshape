package json

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/theory/jsonpath"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

var (
	// ErrNoExpression is returned when the plugin is configured without a
	// JSONPath expression.
	ErrNoExpression = errors.New("json:key requires an 'expression'")
	// ErrInvalidExpression is returned when the expression is not a valid
	// RFC 9535 JSONPath query.
	ErrInvalidExpression = errors.New("invalid JSONPath expression")
	// ErrInvalidJSON is returned when the input cannot be decoded as JSON.
	ErrInvalidJSON = errors.New("invalid JSON input")
)

// Key evaluates an RFC 9535 JSONPath expression against JSON input and emits
// the matched values. It is the JSON counterpart to yaml:key, using the
// JSONPath query language (e.g. "$.a.b[0]", "$.items.*",
// "$.deps[?@.name=='x'].version") rather than a dotted-path lookup.
//
// The emitted data format is derived from the shape of the match, mirroring
// yaml:key: a single scalar emits FormatString, multiple scalars emit
// FormatListString, objects emit FormatMapString or FormatListMapString, and
// no match at all emits FormatNil.
type Key struct {
	fact.BaseFact `yaml:",inline"`

	// Expression is the RFC 9535 JSONPath expression to evaluate against the
	// input.
	Expression string `yaml:"expression"`

	// path holds the expression parsed once and cached, so that a multi-file
	// input does not re-parse it per document.
	path *jsonpath.Path
}

//go:generate go run ../../../cmd/gen.go fact-plugin --package=json

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

// ValidateExpression parses the configured JSONPath expression, returning an
// error if it is missing or malformed. It is safe to call more than once.
//
// Parse is used rather than MustParse so that a malformed user-supplied
// expression surfaces as an error instead of a panic.
func (p *Key) ValidateExpression() error {
	if p.Expression == "" {
		return ErrNoExpression
	}

	if p.path != nil {
		return nil
	}

	parsed, err := jsonpath.Parse(p.Expression)
	if err != nil {
		return fmt.Errorf("%w %q: %s", ErrInvalidExpression, p.Expression, err)
	}

	p.path = parsed
	return nil
}

func (p *Key) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	if err := p.ValidateExpression(); err != nil {
		contextLogger.WithError(err).Error("invalid jsonpath expression")
		p.AddErrors(sentinel(err))
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

		format, values, err := p.query(inputData)
		if err != nil {
			contextLogger.WithError(err).Error("error evaluating jsonpath")
			p.AddErrors(sentinel(err))
			return
		}

		p.Format = format
		p.SetData(values)

	case data.FormatMapBytes:
		inputData := data.AsMapBytes(p.GetInput().GetData())
		if inputData == nil {
			return
		}

		format, values, err := p.queryMap(contextLogger, inputData)
		if err != nil {
			p.AddErrors(sentinel(err))
			return
		}

		p.Format = format
		if format != data.FormatNil {
			p.SetData(values)
		}

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
	for _, s := range []error{ErrNoExpression, ErrInvalidExpression, ErrInvalidJSON} {
		if errors.Is(err, s) {
			return s
		}
	}
	return err
}

// query decodes JSON bytes, evaluates the JSONPath expression, and derives both
// the data format and the value from the shape of the match.
func (p *Key) query(src []byte) (data.DataFormat, interface{}, error) {
	var doc interface{}
	if err := stdjson.Unmarshal(src, &doc); err != nil {
		return data.FormatNil, nil, fmt.Errorf("%w: %s", ErrInvalidJSON, err)
	}

	// SelectLocated is used rather than Select so that results carry their
	// normalised path (e.g. "$['scripts']['lint']"), which Sort orders
	// deterministically. This keeps output stable for expressions that match an
	// unordered JSON object's members via a wildcard, while preserving document
	// order for array matches.
	nodes := p.path.SelectLocated(doc)
	nodes.Sort()

	matches := make([]interface{}, 0, len(nodes))
	for n := range nodes.All() {
		matches = append(matches, n.Node)
	}

	format, value := shapeToData(matches)
	return format, value, nil
}

// fileMatch is a single file's query result, kept alongside its format so that
// queryMap can decide, once every file has been evaluated, whether all files
// share a common shape or need to be promoted to a uniform one.
type fileMatch struct {
	format data.DataFormat
	value  interface{}
}

// queryMap evaluates the expression against every file in a FormatMapBytes
// input and combines the results into a single map-shaped value.
//
// Files are visited in sorted name order and every file's format is recorded
// before any data is combined, so the result does not depend on Go's
// randomised map iteration order: the same input always produces the same
// format and the same data. If every file's match has the same shape, that
// shape is preserved (e.g. a map of strings); if shapes differ across files,
// every value is rendered to a string (or list of strings) so that no file's
// finding is silently dropped.
func (p *Key) queryMap(contextLogger *log.Entry, inputData map[string][]byte) (data.DataFormat, interface{}, error) {
	names := make([]string, 0, len(inputData))
	for name := range inputData {
		names = append(names, name)
	}
	sort.Strings(names)

	matches := map[string]fileMatch{}
	order := make([]string, 0, len(names))
	commonFormat := data.DataFormat("")
	mixed := false

	for _, name := range names {
		format, values, err := p.query(inputData[name])
		if err != nil {
			contextLogger.WithError(err).WithField("file", name).
				Error("error evaluating jsonpath")
			return data.FormatNil, nil, err
		}

		// Skip files where the expression matched nothing, so that a single
		// unmatched file does not affect the shape of the combined result.
		if format == data.FormatNil {
			continue
		}

		matches[name] = fileMatch{format: format, value: values}
		order = append(order, name)

		switch {
		case commonFormat == "":
			commonFormat = format
		case commonFormat != format:
			mixed = true
		}
	}

	if len(order) == 0 {
		return data.FormatNil, nil, nil
	}

	if mixed {
		contextLogger.WithField("files", order).
			Debug("files matched with differing shapes; rendering every file as a list of strings")
		result := make(map[string][]string, len(order))
		for _, name := range order {
			result[name] = toStringList(matches[name].value)
		}
		return data.FormatMapListString, result, nil
	}

	mapFormat := mapValueFormat(commonFormat)
	return mapFormat, buildMapData(mapFormat, order, matches), nil
}

// buildMapData assembles the final map-shaped value once every file is known
// to share the same underlying format. mapFormat is the promoted map format
// (see mapValueFormat); the type assertions below are safe because every
// value in matches was produced by the same commonFormat branch of
// shapeToData.
func buildMapData(mapFormat data.DataFormat, order []string, matches map[string]fileMatch) interface{} {
	switch mapFormat {
	case data.FormatMapString:
		result := make(map[string]string, len(order))
		for _, name := range order {
			result[name] = matches[name].value.(string)
		}
		return result

	case data.FormatMapNestedString:
		result := make(map[string]map[string]string, len(order))
		for _, name := range order {
			result[name] = matches[name].value.(map[string]string)
		}
		return result

	default: // FormatMapListString, promoted from ListString, ListMapString or MapNestedString.
		result := make(map[string][]string, len(order))
		for _, name := range order {
			result[name] = toStringList(matches[name].value)
		}
		return result
	}
}

// toStringList renders any of the shapes shapeToData can produce as a list of
// strings, so that a file's result can always be represented even when its
// siblings matched a different shape.
func toStringList(v interface{}) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []string:
		return t
	case map[string]string:
		return []string{mapToSortedString(t)}
	case []map[string]string:
		out := make([]string, 0, len(t))
		for _, m := range t {
			out = append(out, mapToSortedString(m))
		}
		return out
	case map[string]map[string]string:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		out := make([]string, 0, len(keys))
		for _, k := range keys {
			out = append(out, fmt.Sprintf("%s: %s", k, mapToSortedString(t[k])))
		}
		return out
	default:
		return nil
	}
}

// shapeToData derives the data format and value from the matched nodes,
// mirroring the format selection that yaml:key applies to YAML node kinds.
//
// No match is reported as FormatNil rather than an error, so that analysers
// such as not-empty can act on the absence of a value.
func shapeToData(matches []interface{}) (data.DataFormat, interface{}) {
	if len(matches) == 0 {
		return data.FormatNil, nil
	}

	if len(matches) == 1 {
		switch v := matches[0].(type) {
		case map[string]interface{}:
			if nested, ok := objectOfObjects(v); ok {
				return data.FormatMapNestedString, nested
			}
			return data.FormatMapString, objectToMapString(v)
		case []interface{}:
			// A matched array is emitted as its members, preserving order.
			return listToData(v)
		default:
			return data.FormatString, scalarToString(v)
		}
	}

	return listToData(matches)
}

// listToData renders a list of matches as either a list of strings or, when
// every member is a flat object, a list of maps.
func listToData(matches []interface{}) (data.DataFormat, interface{}) {
	if len(matches) == 0 {
		return data.FormatNil, nil
	}

	allObjects := true
	for _, m := range matches {
		if _, ok := m.(map[string]interface{}); !ok {
			allObjects = false
			break
		}
	}

	if allObjects {
		result := make([]map[string]string, 0, len(matches))
		for _, m := range matches {
			result = append(result, objectToMapString(m.(map[string]interface{})))
		}
		return data.FormatListMapString, result
	}

	result := make([]string, 0, len(matches))
	for _, m := range matches {
		result = append(result, scalarToString(m))
	}
	return data.FormatListString, result
}

// objectOfObjects reports whether every value of an object is itself a flat
// object, and if so returns the nested map representation.
func objectOfObjects(obj map[string]interface{}) (map[string]map[string]string, bool) {
	if len(obj) == 0 {
		return nil, false
	}

	result := map[string]map[string]string{}
	for k, v := range obj {
		inner, ok := v.(map[string]interface{})
		if !ok {
			return nil, false
		}
		result[k] = objectToMapString(inner)
	}
	return result, true
}

// objectToMapString renders a JSON object's values as strings.
func objectToMapString(obj map[string]interface{}) map[string]string {
	result := map[string]string{}
	for k, v := range obj {
		result[k] = scalarToString(v)
	}
	return result
}

// mapValueFormat promotes the format of a per-file value to the format of the
// map that holds it, matching the promotion yaml:key applies across files.
func mapValueFormat(format data.DataFormat) data.DataFormat {
	switch format {
	case data.FormatString:
		return data.FormatMapString
	case data.FormatListString:
		return data.FormatMapListString
	case data.FormatMapString:
		return data.FormatMapNestedString
	default:
		// List-of-map and nested-map results have no map-of-map equivalent in
		// the data package, so fall back to a map of lists of strings.
		return data.FormatMapListString
	}
}

// mapToSortedString renders a map as a deterministic "k=v k2=v2" string, used
// when a map-shaped result has to be flattened into a list. Values are joined
// with strings.Join rather than a slice-typed fmt.Sprintf, which would leak Go's
// "[k=v k2=v2]" bracket syntax into the rendered string.
func mapToSortedString(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return strings.Join(pairs, " ")
}

// scalarToString renders a JSON scalar as a string. Numbers decoded by
// encoding/json are float64; render integral values without a trailing ".0".
//
// A non-scalar value (object or array) reaching this function means a scalar
// was expected but the document held a nested structure instead - e.g. a
// flat-object assumption (objectToMapString) meeting a value that is itself
// an object or array. Rendering it with fmt's default verb would leak Go's
// internal syntax (and the nested structure's contents, which may include
// sensitive values) into breach output, so - mirroring yaml:key, which reads
// a YAML scalar node's Value and gets "" for non-scalar nodes - an empty
// string is returned instead.
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
	case nil:
		return ""
	default:
		return ""
	}
}
