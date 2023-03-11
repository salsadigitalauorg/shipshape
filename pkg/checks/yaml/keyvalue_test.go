package yaml_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/stretchr/testify/assert"
)

func TestMergeKeyValueSlice(t *testing.T) {
	assert := assert.New(t)

	t.Run("emptyBothSlices", func(t *testing.T) {
		slcA := []KeyValue(nil)
		MergeKeyValueSlice(&slcA, []KeyValue(nil))
		assert.Equal([]KeyValue(nil), slcA)
	})

	t.Run("emptyFirstSlice", func(t *testing.T) {
		slcA := []KeyValue(nil)
		MergeKeyValueSlice(&slcA, []KeyValue{
			{Key: "k", Value: "v"},
		})
		assert.Equal([]KeyValue{{Key: "k", Value: "v"}}, slcA)
	})

	t.Run("emptySecondSlice", func(t *testing.T) {
		slcA := []KeyValue{{Key: "k", Value: "v"}}
		MergeKeyValueSlice(&slcA, []KeyValue(nil))
		assert.Equal([]KeyValue{{Key: "k", Value: "v"}}, slcA)

		MergeKeyValueSlice(&slcA, []KeyValue{})
		assert.Equal([]KeyValue{{Key: "k", Value: "v"}}, slcA)
	})

	t.Run("valuesInBoth", func(t *testing.T) {
		slcA := []KeyValue{
			{Key: "k1", Value: "v1"},
		}
		MergeKeyValueSlice(&slcA, []KeyValue{
			{Key: "k2", Value: "v2"},
		})
		assert.Equal([]KeyValue{{Key: "k2", Value: "v2"}}, slcA)
	})
}

func TestEquals(t *testing.T) {
	assert := assert.New(t)

	t.Run("stringEquals", func(t *testing.T) {
		kv := KeyValue{Key: "k", Value: "foo"}
		assert.True(kv.Equals("foo"))
		assert.False(kv.Equals("bar"))
	})

	t.Run("stringEqualsTruthyTrue", func(t *testing.T) {
		kv := KeyValue{Key: "k", Value: "true", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("false"))
		assert.False(kv.Equals("0"))
		assert.False(kv.Equals("null"))
		assert.True(kv.Equals("true"))
		assert.True(kv.Equals("1"))
	})

	t.Run("stringEqualsTruthyOne", func(t *testing.T) {
		kv := KeyValue{Key: "k", Value: "1", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("false"))
		assert.False(kv.Equals("0"))
		assert.False(kv.Equals("null"))
		assert.True(kv.Equals("true"))
		assert.True(kv.Equals("1"))
	})

	t.Run("stringEqualsTruthyFalse", func(t *testing.T) {
		kv := KeyValue{Key: "k", Value: "false", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("true"))
		assert.False(kv.Equals("1"))
		assert.True(kv.Equals("null"))
		assert.True(kv.Equals("false"))
		assert.True(kv.Equals("0"))
	})

	t.Run("stringEqualsTruthyZero", func(t *testing.T) {
		kv := KeyValue{Key: "k", Value: "0", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("true"))
		assert.False(kv.Equals("1"))
		assert.True(kv.Equals("null"))
		assert.True(kv.Equals("false"))
		assert.True(kv.Equals("0"))
	})

	t.Run("stringEqualsTruthyNull", func(t *testing.T) {
		kv := KeyValue{Key: "k", Value: "null", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("true"))
		assert.False(kv.Equals("1"))
		assert.True(kv.Equals("null"))
		assert.True(kv.Equals("false"))
		assert.True(kv.Equals("0"))
	})
}
