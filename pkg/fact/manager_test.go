package fact_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/fact/testdata"
)

func registerFact(name string, format data.DataFormat, d any) {
	f := testdata.New(name, format, d)
	f.Collect()
	fact.Manager().SetPlugins(map[string]fact.Facter{name: f})
}

func TestLookupFactAsStringMap(t *testing.T) {
	defer fact.Manager().ResetPlugins()

	tt := []struct {
		name      string
		format    data.DataFormat
		data      any
		inputName string
		key       string
		expected  string
	}{
		{
			name:      "mapStringHit",
			format:    data.FormatMapString,
			data:      map[string]string{"contact_form": "Contact us"},
			inputName: "titles",
			key:       "contact_form",
			expected:  "Contact us",
		},
		{
			name:      "mapStringMissingKey",
			format:    data.FormatMapString,
			data:      map[string]string{"contact_form": "Contact us"},
			inputName: "titles",
			key:       "does_not_exist",
			expected:  "",
		},
		{
			name:      "mapNestedStringDoesNotPanic",
			format:    data.FormatMapNestedString,
			data:      map[string]map[string]string{"contact_form": {"title": "Contact us"}},
			inputName: "titles",
			key:       "contact_form",
			expected:  "",
		},
		{
			name:      "unsupportedFormatDoesNotPanic",
			format:    data.FormatListString,
			data:      []string{"a", "b"},
			inputName: "titles",
			key:       "a",
			expected:  "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			fact.Manager().ResetPlugins()
			registerFact(tc.inputName, tc.format, tc.data)

			assert.NotPanics(t, func() {
				got := fact.LookupFactAsStringMap(tc.inputName, tc.key)
				assert.Equal(t, tc.expected, got)
			})
		})
	}
}

func TestLookupFactAsStringMapUnknownInput(t *testing.T) {
	defer fact.Manager().ResetPlugins()
	fact.Manager().ResetPlugins()

	assert.NotPanics(t, func() {
		assert.Equal(t, "", fact.LookupFactAsStringMap("nonexistent", "key"))
	})
}
