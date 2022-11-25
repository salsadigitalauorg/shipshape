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
