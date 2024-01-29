package result_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/result"

	"github.com/stretchr/testify/assert"
)

func TestResultSort(t *testing.T) {
	assert := assert.New(t)

	r := Result{
		Passes: []string{"z pass", "g pass", "a pass", "b pass"},
		Breaches: []Breach{
			ValueBreach{CheckName: "x", Value: "breach 1"},
			ValueBreach{CheckName: "h", Value: "breach 2"},
			ValueBreach{CheckName: "v", Value: "breach 3"},
			ValueBreach{CheckName: "f", Value: "breach 4"},
		},
		Warnings: []string{"y warn", "i warn", "u warn", "c warn"},
	}
	r.Sort()

	assert.EqualValues(Result{
		Passes: []string{"a pass", "b pass", "g pass", "z pass"},
		Breaches: []Breach{
			ValueBreach{CheckName: "f", Value: "breach 4"},
			ValueBreach{CheckName: "h", Value: "breach 2"},
			ValueBreach{CheckName: "v", Value: "breach 3"},
			ValueBreach{CheckName: "x", Value: "breach 1"},
		},
		Warnings: []string{"c warn", "i warn", "u warn", "y warn"},
	}, r)
}
