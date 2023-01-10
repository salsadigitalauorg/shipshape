package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestMergeKeyValueSlice(t *testing.T) {
	assert := assert.New(t)

	t.Run("emptyBothSlices", func(t *testing.T) {
		slcA := []shipshape.KeyValue(nil)
		shipshape.MergeKeyValueSlice(&slcA, []shipshape.KeyValue(nil))
		assert.Equal([]shipshape.KeyValue(nil), slcA)
	})

	t.Run("emptyFirstSlice", func(t *testing.T) {
		slcA := []shipshape.KeyValue(nil)
		shipshape.MergeKeyValueSlice(&slcA, []shipshape.KeyValue{
			{Key: "k", Value: "v"},
		})
		assert.Equal([]shipshape.KeyValue{{Key: "k", Value: "v"}}, slcA)
	})

	t.Run("emptySecondSlice", func(t *testing.T) {
		slcA := []shipshape.KeyValue{{Key: "k", Value: "v"}}
		shipshape.MergeKeyValueSlice(&slcA, []shipshape.KeyValue(nil))
		assert.Equal([]shipshape.KeyValue{{Key: "k", Value: "v"}}, slcA)

		shipshape.MergeKeyValueSlice(&slcA, []shipshape.KeyValue{})
		assert.Equal([]shipshape.KeyValue{{Key: "k", Value: "v"}}, slcA)
	})

	t.Run("valuesInBoth", func(t *testing.T) {
		slcA := []shipshape.KeyValue{
			{Key: "k1", Value: "v1"},
		}
		shipshape.MergeKeyValueSlice(&slcA, []shipshape.KeyValue{
			{Key: "k2", Value: "v2"},
		})
		assert.Equal([]shipshape.KeyValue{{Key: "k2", Value: "v2"}}, slcA)
	})
}

func TestEquals(t *testing.T) {
	assert := assert.New(t)

	t.Run("stringEquals", func(t *testing.T) {
		kv := shipshape.KeyValue{Key: "k", Value: "foo"}
		assert.True(kv.Equals("foo"))
		assert.False(kv.Equals("bar"))
	})

	t.Run("stringEqualsTruthyTrue", func(t *testing.T) {
		kv := shipshape.KeyValue{Key: "k", Value: "true", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("false"))
		assert.False(kv.Equals("0"))
		assert.False(kv.Equals("null"))
		assert.True(kv.Equals("true"))
		assert.True(kv.Equals("1"))
	})

	t.Run("stringEqualsTruthyOne", func(t *testing.T) {
		kv := shipshape.KeyValue{Key: "k", Value: "1", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("false"))
		assert.False(kv.Equals("0"))
		assert.False(kv.Equals("null"))
		assert.True(kv.Equals("true"))
		assert.True(kv.Equals("1"))
	})

	t.Run("stringEqualsTruthyFalse", func(t *testing.T) {
		kv := shipshape.KeyValue{Key: "k", Value: "false", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("true"))
		assert.False(kv.Equals("1"))
		assert.True(kv.Equals("null"))
		assert.True(kv.Equals("false"))
		assert.True(kv.Equals("0"))
	})

	t.Run("stringEqualsTruthyZero", func(t *testing.T) {
		kv := shipshape.KeyValue{Key: "k", Value: "0", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("true"))
		assert.False(kv.Equals("1"))
		assert.True(kv.Equals("null"))
		assert.True(kv.Equals("false"))
		assert.True(kv.Equals("0"))
	})

	t.Run("stringEqualsTruthyNull", func(t *testing.T) {
		kv := shipshape.KeyValue{Key: "k", Value: "null", Truthy: true}
		assert.False(kv.Equals("foo"))
		assert.False(kv.Equals("true"))
		assert.False(kv.Equals("1"))
		assert.True(kv.Equals("null"))
		assert.True(kv.Equals("false"))
		assert.True(kv.Equals("0"))
	})
}
