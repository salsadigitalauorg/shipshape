package shipshape_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestResultSort(t *testing.T) {
	assert := assert.New(t)

	r := Result{
		Passes:   []string{"z pass", "g pass", "a pass", "b pass"},
		Failures: []string{"x fail", "h fail", "v fail", "f fail"},
		Warnings: []string{"y warn", "i warn", "u warn", "c warn"},
	}
	r.Sort()

	assert.EqualValues(Result{
		Passes:   []string{"a pass", "b pass", "g pass", "z pass"},
		Failures: []string{"f fail", "h fail", "v fail", "x fail"},
		Warnings: []string{"c warn", "i warn", "u warn", "y warn"},
	}, r)
}
