package fact_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/fact/testdata"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
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

// dependentTestFacter builds a TestFacter that declares a required input of
// data.FormatString, so ValidateInput() actually passes and Collect() is
// reached - without this, every TestFacter defaults to SupportNone/no
// formats and ValidateInput() rejects the fact before the code path under
// test (CollectFact's error propagation) ever runs, making the regression
// test pass regardless of whether the bug is present.
func dependentTestFacter(id string) *testdata.TestFacter {
	d := testdata.New(id, data.FormatString, "should not collect")
	d.TestSupportedInputFormatsLevel = plugin.SupportRequired
	d.TestSupportedInputFormatsList = []data.DataFormat{data.FormatString}
	return d
}

// TestCollectFactPrimaryInputErrorStopsCollection ensures a failure in the
// *primary* input still halts collection of the dependent fact, even when an
// additional input is also configured and collects successfully. Regression
// test for a bug where CollectFact reused a single loop variable across the
// primary-input and additional-inputs loops, so the error check after both
// loops only ever inspected the last additional input.
func TestCollectFactPrimaryInputErrorStopsCollection(t *testing.T) {
	defer fact.Manager().ResetPlugins()
	fact.Manager().ResetPlugins()

	primary := testdata.New("primary", data.FormatString, nil)
	primary.TestError = errors.New("primary input failed")

	additional := testdata.New("additional", data.FormatString, "ok")

	dependent := dependentTestFacter("dependent")
	dependent.SetInputName("primary")
	dependent.AdditionalInputNames = []string{"additional"}

	fact.Manager().SetPlugins(map[string]fact.Facter{
		"primary":    primary,
		"additional": additional,
		"dependent":  dependent,
	})

	fact.Manager().CollectFact("dependent", dependent)

	// Sanity check: confirm the primary input's error was actually recorded,
	// so a passing assertion below can't be explained by the primary
	// silently succeeding instead of by the fix working.
	require.NotEmpty(t, primary.GetErrors(), "primary input must have recorded its error")

	assert.False(t, dependent.TestCollected,
		"dependent fact's Collect() must not have run when its primary input errored")
}

// TestCollectFactAdditionalInputErrorStopsCollection is the mirror case:
// a failure in an *earlier* additional input must halt collection even when
// a *later* additional input succeeds. This shape specifically targets the
// original bug: CollectFact aliased a single loop variable across every
// additional input, so after the loop it only ever reflected the last one
// processed - a failure earlier in the list was silently overwritten by a
// later success. A single-additional-input case (as this test previously
// used) cannot distinguish the fix from the bug, because with only one
// additional input, "last processed" and "the erroring one" are always the
// same input.
func TestCollectFactAdditionalInputErrorStopsCollection(t *testing.T) {
	defer fact.Manager().ResetPlugins()
	fact.Manager().ResetPlugins()

	primary := testdata.New("primary", data.FormatString, "ok")

	additionalFailing := testdata.New("additional-failing", data.FormatString, nil)
	additionalFailing.TestError = errors.New("additional input failed")

	additionalOK := testdata.New("additional-ok", data.FormatString, "ok")

	dependent := dependentTestFacter("dependent")
	dependent.SetInputName("primary")
	// additional-failing is deliberately listed before additional-ok, so a
	// manager that only inspects the last additional input processed would
	// see the (later, successful) additional-ok and incorrectly conclude
	// nothing failed.
	dependent.AdditionalInputNames = []string{"additional-failing", "additional-ok"}

	fact.Manager().SetPlugins(map[string]fact.Facter{
		"primary":            primary,
		"additional-failing": additionalFailing,
		"additional-ok":      additionalOK,
		"dependent":          dependent,
	})

	fact.Manager().CollectFact("dependent", dependent)

	require.NotEmpty(t, additionalFailing.GetErrors(), "additional-failing input must have recorded its error")
	require.Empty(t, additionalOK.GetErrors(), "additional-ok input must not have errored")

	assert.False(t, dependent.TestCollected,
		"dependent fact's Collect() must not have run when an earlier additional input errored, "+
			"even though a later additional input succeeded")
}

// TestCollectFactCollectsWhenNoInputErrors is the converse check: with both
// the primary and additional input free of errors, the dependent fact's
// Collect() must actually run. Without this, TestCollectFact*StopsCollection
// above could be satisfied by a manager that never calls Collect() at all.
func TestCollectFactCollectsWhenNoInputErrors(t *testing.T) {
	defer fact.Manager().ResetPlugins()
	fact.Manager().ResetPlugins()

	primary := testdata.New("primary", data.FormatString, "ok")
	additional := testdata.New("additional", data.FormatString, "ok")

	dependent := dependentTestFacter("dependent")
	dependent.SetInputName("primary")
	dependent.AdditionalInputNames = []string{"additional"}

	fact.Manager().SetPlugins(map[string]fact.Facter{
		"primary":    primary,
		"additional": additional,
		"dependent":  dependent,
	})

	fact.Manager().CollectFact("dependent", dependent)

	assert.Empty(t, dependent.GetErrors())
	assert.True(t, dependent.TestCollected,
		"dependent fact's Collect() must have run when neither input errored")
	assert.Equal(t, data.FormatString, dependent.GetFormat(),
		"dependent fact's Collect() must have run and set its format/data")
}
