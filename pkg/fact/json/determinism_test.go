package json

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
)

// TestObjectWildcardDeterminism asserts that matching the members of an
// (unordered) JSON object yields a stable order across repeated evaluations.
// Go randomises map iteration, so this fails without the normalised-path sort
// in query().
func TestObjectWildcardDeterminism(t *testing.T) {
	assert := assert.New(t)

	src := []byte(`{"scripts":{"lint":"eslint .","build":"vite build",` +
		`"test":"vitest run","fmt":"prettier -w .","e2e":"playwright test"}}`)

	var first []string
	for i := 0; i < 25; i++ {
		p := New("json-values")
		p.Expression = "$.scripts.*"
		assert.NoError(p.ValidateExpression())

		format, values, err := p.query(src)
		assert.NoError(err)
		assert.Equal(data.FormatListString, format)

		got, ok := values.([]string)
		assert.True(ok)

		if first == nil {
			first = got
			continue
		}
		assert.Equal(first, got, "object wildcard order changed between runs")
	}

	// Sorted by normalised path, i.e. by member name: build, e2e, fmt, lint, test.
	assert.Equal([]string{
		"vite build", "playwright test", "prettier -w .", "eslint .", "vitest run",
	}, first)
}

// TestArrayOrderPreserved asserts document order is retained for array matches,
// which a blanket sort of the values would break.
func TestArrayOrderPreserved(t *testing.T) {
	assert := assert.New(t)

	p := New("json-values")
	p.Expression = "$.items[*]"
	assert.NoError(p.ValidateExpression())

	format, values, err := p.query([]byte(`{"items":["c","a","b","10","2"]}`))
	assert.NoError(err)
	assert.Equal(data.FormatListString, format)
	assert.Equal([]string{"c", "a", "b", "10", "2"}, values)
}

// TestMultiFileMixedShapesDeterminism is a regression test for a bug where
// combining multiple files' JSONPath matches into one map-shaped fact depended
// on Go's randomised map iteration order: whichever file was visited first
// fixed the result's format, and every file that did not share that shape was
// silently dropped from the output. Against two files whose matches have
// different shapes (one object, one scalar), this asserts every run over many
// iterations produces the same format and retains both files' data.
func TestMultiFileMixedShapesDeterminism(t *testing.T) {
	assert := assert.New(t)

	inputData := map[string][]byte{
		"x.json": []byte(`{"a":{"k":"v"}}`),
		"y.json": []byte(`{"a":"scalarValue"}`),
	}
	logger := log.WithField("test", "TestMultiFileMixedShapesDeterminism")

	for i := 0; i < 100; i++ {
		p := New("json-values")
		p.Expression = "$.a"
		assert.NoError(p.ValidateExpression())

		format, values, err := p.queryMap(logger, inputData)
		assert.NoError(err)
		assert.Equal(data.FormatMapListString, format,
			"run %d: format changed between runs", i)

		got, ok := values.(map[string][]string)
		assert.True(ok, "run %d: unexpected value type %T", i, values)
		assert.Len(got, 2, "run %d: a file's match was dropped", i)
		assert.Equal([]string{"k=v"}, got["x.json"], "run %d", i)
		assert.Equal([]string{"scalarValue"}, got["y.json"], "run %d", i)
	}
}

// TestMultiFileUniformShapeDeterminism covers the common case - every file's
// match shares the same shape - across many runs, since map iteration order
// could in principle affect which file's error surfaces first even when the
// combined format itself is stable.
func TestMultiFileUniformShapeDeterminism(t *testing.T) {
	assert := assert.New(t)

	inputData := map[string][]byte{
		"a.json": []byte(`{"name":"alpha"}`),
		"b.json": []byte(`{"name":"beta"}`),
		"c.json": []byte(`{"name":"gamma"}`),
	}
	logger := log.WithField("test", "TestMultiFileUniformShapeDeterminism")

	want := map[string]string{"a.json": "alpha", "b.json": "beta", "c.json": "gamma"}

	for i := 0; i < 100; i++ {
		p := New("json-values")
		p.Expression = "$.name"
		assert.NoError(p.ValidateExpression())

		format, values, err := p.queryMap(logger, inputData)
		assert.NoError(err)
		assert.Equal(data.FormatMapString, format, "run %d", i)
		assert.Equal(want, values, "run %d", i)
	}
}
