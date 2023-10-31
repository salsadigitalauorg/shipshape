package json_test

import (
	. "github.com/salsadigitalauorg/shipshape/pkg/checks/json"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsEmpty(t *testing.T) {
	assertions := assert.New(t)

	kv := KeyValue{}
	assertions.True(kv.IsEmpty(nil))
	assertions.True(kv.IsEmpty(""))
	assertions.True(kv.IsEmpty([]any{}))
	assertions.True(kv.IsEmpty([]string{}))
	assertions.True(kv.IsEmpty([]int{}))
	assertions.True(kv.IsEmpty([0]int{}))
	assertions.True(kv.IsEmpty(map[string]string{}))
	assertions.True(kv.IsEmpty(map[any]any{}))

	assertions.False(kv.IsEmpty("foo"))
	assertions.False(kv.IsEmpty(false))
	assertions.False(kv.IsEmpty(true))
	assertions.False(kv.IsEmpty(0))
	assertions.False(kv.IsEmpty([]any{0, nil, "", false}))
	assertions.False(kv.IsEmpty([]int{0, 1}))
	assertions.False(kv.IsEmpty([]string{"foo", "bar"}))
	assertions.False(kv.IsEmpty(map[any]any{
		0:      0,
		"0":    1,
		1:      "1",
		"true": true,
		false:  "false",
		true:   false,
		"foo":  "bar",
	}))
}

func TestEquals(t *testing.T) {
	assertions := assert.New(t)

	t.Run("stringEquals", func(t *testing.T) {
		kv := KeyValue{KeyValue: yaml.KeyValue{Key: "k", Value: "foo"}}
		assertions.True(kv.Equals("foo"))
		assertions.True(kv.Equals("FoO"))
		assertions.False(kv.Equals("bar"))
	})

	t.Run("stringEqualsTruthyTrue", func(t *testing.T) {
		kv := KeyValue{KeyValue: yaml.KeyValue{Key: "k", Value: "true", Truthy: true}}
		assertions.False(kv.Equals("foo"))
		assertions.False(kv.Equals("false"))
		assertions.False(kv.Equals("0"))
		assertions.False(kv.Equals("null"))
		assertions.True(kv.Equals("true"))
		assertions.True(kv.Equals("1"))
	})

	t.Run("stringEqualsTruthyOne", func(t *testing.T) {
		kv := KeyValue{KeyValue: yaml.KeyValue{Key: "k", Value: "1", Truthy: true}}
		assertions.False(kv.Equals("foo"))
		assertions.False(kv.Equals("false"))
		assertions.False(kv.Equals("0"))
		assertions.False(kv.Equals("null"))
		assertions.True(kv.Equals("true"))
		assertions.True(kv.Equals("1"))
	})

	t.Run("stringEqualsTruthyFalse", func(t *testing.T) {
		kv := KeyValue{KeyValue: yaml.KeyValue{Key: "k", Value: "false", Truthy: true}}
		assertions.False(kv.Equals("foo"))
		assertions.False(kv.Equals("true"))
		assertions.False(kv.Equals("1"))
		assertions.True(kv.Equals("null"))
		assertions.True(kv.Equals("false"))
		assertions.True(kv.Equals("0"))
	})

	t.Run("stringEqualsTruthyZero", func(t *testing.T) {
		kv := KeyValue{KeyValue: yaml.KeyValue{Key: "k", Value: "0", Truthy: true}}
		assertions.False(kv.Equals("foo"))
		assertions.False(kv.Equals("true"))
		assertions.False(kv.Equals("1"))
		assertions.True(kv.Equals("null"))
		assertions.True(kv.Equals("false"))
		assertions.True(kv.Equals("0"))
	})

	t.Run("stringEqualsTruthyNull", func(t *testing.T) {
		kv := KeyValue{KeyValue: yaml.KeyValue{Key: "k", Value: "null", Truthy: true}}
		assertions.False(kv.Equals("foo"))
		assertions.False(kv.Equals("true"))
		assertions.False(kv.Equals("1"))
		assertions.True(kv.Equals("null"))
		assertions.True(kv.Equals("false"))
		assertions.True(kv.Equals("0"))
	})

	t.Run("stringEqualsAny", func(t *testing.T) {
		kv := KeyValue{KeyValue: yaml.KeyValue{Key: "k", Value: "foo"}}
		assertions.True(kv.Equals("foo"))
		assertions.True(kv.Equals("FoO"))
		assertions.False(kv.Equals("bar"))
		assertions.False(kv.Equals(true))
		assertions.False(kv.Equals(1))
		assertions.False(kv.Equals(nil))
	})

	t.Run("stringEqualsTruthyTrueAny", func(t *testing.T) {
		kv := KeyValue{KeyValue: yaml.KeyValue{Key: "k", Value: "true", Truthy: true}}
		assertions.False(kv.Equals("foo"))
		assertions.True(kv.Equals("true"))
		assertions.True(kv.Equals(true))
		assertions.True(kv.Equals("1"))
		assertions.True(kv.Equals(1))
		assertions.False(kv.Equals("null"))
		assertions.False(kv.Equals("false"))
		assertions.False(kv.Equals(false))
		assertions.False(kv.Equals("0"))
		assertions.False(kv.Equals(0))
	})

	t.Run("stringEqualsTruthyFalseAny", func(t *testing.T) {
		kv := KeyValue{KeyValue: yaml.KeyValue{Key: "k", Value: "false", Truthy: true}}
		assertions.False(kv.Equals("foo"))
		assertions.False(kv.Equals("true"))
		assertions.False(kv.Equals(true))
		assertions.False(kv.Equals("1"))
		assertions.False(kv.Equals(1))
		assertions.True(kv.Equals("null"))
		assertions.True(kv.Equals("false"))
		assertions.True(kv.Equals(false))
		assertions.True(kv.Equals("0"))
		assertions.True(kv.Equals(0))
	})
}

func TestIsDisallowed(t *testing.T) {
	assertions := assert.New(t)

	t.Run("testDisallowedList", func(t *testing.T) {
		kv := KeyValue{DisallowedValues: []any{true, 1, "false", "foo"}}
		assertions.False(kv.IsDisallowed(""))
		assertions.False(kv.IsDisallowed(nil))
		assertions.True(kv.IsDisallowed("foo"))
		assertions.False(kv.IsDisallowed("FOo"))
		assertions.False(kv.IsDisallowed("bar"))
		assertions.False(kv.IsDisallowed("1"))
		assertions.True(kv.IsDisallowed(1))
		assertions.True(kv.IsDisallowed(true))
		assertions.False(kv.IsDisallowed("true"))
		assertions.False(kv.IsDisallowed(false))
		assertions.True(kv.IsDisallowed("false"))
	})

	t.Run("testAllowedList", func(t *testing.T) {
		kv := KeyValue{AllowedValues: []any{true, 1, "false", "foo"}}
		assertions.False(kv.IsDisallowed(""))
		assertions.False(kv.IsDisallowed(nil))
		assertions.False(kv.IsDisallowed("foo"))
		assertions.True(kv.IsDisallowed("FOo"))
		assertions.True(kv.IsDisallowed("bar"))
		assertions.True(kv.IsDisallowed("1"))
		assertions.False(kv.IsDisallowed(1))
		assertions.False(kv.IsDisallowed(true))
		assertions.True(kv.IsDisallowed("true"))
		assertions.True(kv.IsDisallowed(false))
		assertions.False(kv.IsDisallowed("false"))
	})

	t.Run("testBothList", func(t *testing.T) {
		kv := KeyValue{
			AllowedValues:    []any{true, 1, "false", "foo"},
			DisallowedValues: []any{false, 0, "true", "bar"},
		}
		assertions.False(kv.IsDisallowed(""))
		assertions.False(kv.IsDisallowed(nil))
		assertions.False(kv.IsDisallowed("foo"))
		assertions.True(kv.IsDisallowed("FOo"))
		assertions.True(kv.IsDisallowed("bar"))
		assertions.True(kv.IsDisallowed("foobar"))
		assertions.True(kv.IsDisallowed("1"))
		assertions.False(kv.IsDisallowed(1))
		assertions.True(kv.IsDisallowed("0"))
		assertions.True(kv.IsDisallowed(0))
		assertions.False(kv.IsDisallowed(true))
		assertions.True(kv.IsDisallowed("true"))
		assertions.True(kv.IsDisallowed(false))
		assertions.False(kv.IsDisallowed("false"))
	})
}
