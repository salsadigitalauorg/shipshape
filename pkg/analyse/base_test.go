package analyse_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/analyse"
)

func TestBaseAnalyserGetResultSource(t *testing.T) {
	t.Run("source set propagates to Result", func(t *testing.T) {
		p := BaseAnalyser{Source: "shipshape:custom:source"}
		r := p.GetResult()
		assert.Equal(t, "shipshape:custom:source", r.Source)
	})

	t.Run("empty source leaves Result.Source empty", func(t *testing.T) {
		p := BaseAnalyser{}
		r := p.GetResult()
		assert.Equal(t, "", r.Source)
	})
}

func TestBaseAnalyserGetSource(t *testing.T) {
	p := BaseAnalyser{Source: "shipshape:custom:source"}
	assert.Equal(t, "shipshape:custom:source", p.GetSource())
}
