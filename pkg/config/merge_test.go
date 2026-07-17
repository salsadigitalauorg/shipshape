package config

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDeepMerge(t *testing.T) {
	currLogOut := logrus.StandardLogger().Out
	defer logrus.SetOutput(currLogOut)
	logrus.SetOutput(io.Discard)

	t.Run("disjointKeysUnion", func(t *testing.T) {
		assert := assert.New(t)
		dst := map[string]interface{}{"a": 1}
		src := map[string]interface{}{"b": 2}
		got := deepMerge(dst, src, "", "file 2")
		assert.Equal(map[string]interface{}{"a": 1, "b": 2}, got)
	})

	t.Run("nestedMapRecursesLeafOnly", func(t *testing.T) {
		assert := assert.New(t)
		dst := map[string]interface{}{
			"plugin": map[string]interface{}{"threshold": 10, "name": "keep"},
		}
		src := map[string]interface{}{
			"plugin": map[string]interface{}{"threshold": 20},
		}
		got := deepMerge(dst, src, "", "file 2")
		assert.Equal(map[string]interface{}{
			"plugin": map[string]interface{}{"threshold": 20, "name": "keep"},
		}, got)
	})

	t.Run("scalarOverrideReplaces", func(t *testing.T) {
		assert := assert.New(t)
		dst := map[string]interface{}{"a": 1}
		src := map[string]interface{}{"a": 2}
		got := deepMerge(dst, src, "", "file 2")
		assert.Equal(map[string]interface{}{"a": 2}, got)
	})

	t.Run("sliceOverrideReplacesNotAppends", func(t *testing.T) {
		assert := assert.New(t)
		dst := map[string]interface{}{"ignore": []interface{}{"a", "b"}}
		src := map[string]interface{}{"ignore": []interface{}{"c"}}
		got := deepMerge(dst, src, "", "file 2")
		assert.Equal(map[string]interface{}{"ignore": []interface{}{"c"}}, got)
	})

	t.Run("mapReplacesScalarWhenTypesDiffer", func(t *testing.T) {
		assert := assert.New(t)
		dst := map[string]interface{}{"a": 1}
		src := map[string]interface{}{"a": map[string]interface{}{"nested": true}}
		got := deepMerge(dst, src, "", "file 2")
		assert.Equal(map[string]interface{}{
			"a": map[string]interface{}{"nested": true},
		}, got)
	})

	t.Run("nilDstInitialises", func(t *testing.T) {
		assert := assert.New(t)
		src := map[string]interface{}{"a": 1}
		got := deepMerge(nil, src, "", "file 1")
		assert.Equal(map[string]interface{}{"a": 1}, got)
	})

	t.Run("emptySrcNoChange", func(t *testing.T) {
		assert := assert.New(t)
		dst := map[string]interface{}{"a": 1}
		got := deepMerge(dst, map[string]interface{}{}, "", "file 2")
		assert.Equal(map[string]interface{}{"a": 1}, got)
	})
}

func TestHasAnyV2Key(t *testing.T) {
	assert := assert.New(t)

	assert.True(hasAnyV2Key(map[string]interface{}{"collect": map[string]interface{}{}}))
	assert.True(hasAnyV2Key(map[string]interface{}{"analyse": nil}))
	assert.True(hasAnyV2Key(map[string]interface{}{"connections": nil}))
	assert.True(hasAnyV2Key(map[string]interface{}{"output": nil}))
	assert.False(hasAnyV2Key(map[string]interface{}{"checks": nil}))
	assert.False(hasAnyV2Key(map[string]interface{}{}))
}
