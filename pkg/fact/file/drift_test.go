package file_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/file"
	"github.com/salsadigitalauorg/shipshape/pkg/fact/testdata"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

func TestDriftInit(t *testing.T) {
	assert := assert.New(t)

	factPlugin := fact.Manager().GetFactories()["file:drift"]("testDrift")
	assert.NotNil(factPlugin)
	driftFacter, ok := factPlugin.(*Drift)
	assert.True(ok)
	assert.Equal("testDrift", driftFacter.GetId())
}

func TestDriftPluginName(t *testing.T) {
	assert.Equal(t, "file:drift", NewDrift("testDrift").GetName())
}

func TestDriftSupportedInputFormats(t *testing.T) {
	d := NewDrift("testDrift")
	supportLevel, inputFormats := d.SupportedInputFormats()
	assert.Equal(t, plugin.SupportRequired, supportLevel)
	assert.ElementsMatch(t, []data.DataFormat{
		data.FormatRaw,
		data.FormatString,
	}, inputFormats)
}

func TestDriftCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name:               "noInput",
			Facter:             NewDrift("drift"),
			ExpectedInputError: &plugin.ErrSupportRequired{SupportType: "input"},
		},
		{
			Name: "noAdditionalInputs",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("runs-on: ubuntu-latest\n"),
			},
			ExpectedErrors: []error{ErrNoCurrentInput},
		},
		{
			Name: "identicalNoPlaceholders",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				f.AdditionalInputNames = []string{"current"}
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte("runs-on: ubuntu-latest\nsteps:\n  - run: build\n"),
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"current": {
					DataFormat: data.FormatRaw,
					Data:       []byte("runs-on: ubuntu-latest\nsteps:\n  - run: build\n"),
				},
			},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{},
		},
		{
			Name: "placeholderSubstitutedNoDrift",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				f.AdditionalInputNames = []string{"current"}
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte("name: {{ PROJECT_NAME }}\nruns-on: ubuntu-latest\n"),
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"current": {
					DataFormat: data.FormatRaw,
					Data:       []byte("name: my-actual-project\nruns-on: ubuntu-latest\n"),
				},
			},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{},
		},
		{
			Name: "githubActionsExpressionNotTreatedAsDrift",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				f.AdditionalInputNames = []string{"current"}
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte("token: ${{ secrets.GITHUB_TOKEN }}\n"),
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"current": {
					DataFormat: data.FormatRaw,
					Data:       []byte("token: ${{ secrets.GITHUB_TOKEN }}\n"),
				},
			},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{},
		},
		{
			Name: "genuineDriftProducesDiff",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				f.AdditionalInputNames = []string{"current"}
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte("name: {{ PROJECT_NAME }}\nruns-on: ubuntu-latest\n"),
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"current": {
					DataFormat: data.FormatRaw,
					Data:       []byte("name: my-actual-project\nruns-on: ubuntu-20.04\n"),
				},
			},
			ExpectedFormat: data.FormatListString,
			// Non-empty; exact diff content is asserted in a dedicated test
			// below since ElementsMatch on diff lines is order-sensitive.
		},
		{
			Name: "unsubstitutedPlaceholderFlagged",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				f.AdditionalInputNames = []string{"current"}
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte("name: {{ PROJECT_NAME }}\n"),
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"current": {
					DataFormat: data.FormatRaw,
					Data:       []byte("name: {{ PROJECT_NAME }}\n"),
				},
			},
			ExpectedFormat: data.FormatListString,
			// Non-empty; asserted in a dedicated test below.
		},
		{
			Name: "invalidPlaceholderPattern",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				f.AdditionalInputNames = []string{"current"}
				f.PlaceholderPattern = "(unclosed"
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw, Data: []byte("name: x\n"),
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"current": {DataFormat: data.FormatRaw, Data: []byte("name: x\n")},
			},
			ExpectedErrors: []error{ErrInvalidPlaceholderPattern},
		},
		{
			// Regression test: an additional input in a format file:drift
			// cannot read (e.g. map-bytes from file:lookup, or list-string
			// from json:key) must error, not silently stringify to "" and
			// produce a false-positive breach claiming the entire template
			// was deleted.
			Name: "unsupportedAdditionalInputFormatMapBytes",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				f.AdditionalInputNames = []string{"current"}
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte("name: {{ P }}\nkeep: 1\n"),
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"current": {
					DataFormat: data.FormatMapBytes,
					Data:       map[string][]byte{"a": []byte("x")},
				},
			},
			ExpectedErrors: []error{&plugin.ErrSupportNone{
				Plugin:        "file:drift",
				SupportType:   "additionalInputFormat",
				SupportPlugin: string(data.FormatMapBytes),
			}},
		},
		{
			Name: "unsupportedAdditionalInputFormatListString",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				f.AdditionalInputNames = []string{"current"}
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte("name: {{ P }}\nkeep: 1\n"),
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"current": {
					DataFormat: data.FormatListString,
					Data:       []string{"x"},
				},
			},
			ExpectedErrors: []error{&plugin.ErrSupportNone{
				Plugin:        "file:drift",
				SupportType:   "additionalInputFormat",
				SupportPlugin: string(data.FormatListString),
			}},
		},
		{
			Name: "customPlaceholderPattern",
			FactFn: func() fact.Facter {
				f := NewDrift("drift")
				f.SetInputName("test-input")
				f.AdditionalInputNames = []string{"current"}
				f.PlaceholderPattern = `__(\w+)__`
				return f
			},
			TestInput: internal.FactInputTest{
				DataFormat: data.FormatRaw,
				Data:       []byte("name: __PROJECT_NAME__\n"),
			},
			TestAdditionalInputs: map[string]internal.FactInputTest{
				"current": {
					DataFormat: data.FormatRaw,
					Data:       []byte("name: my-actual-project\n"),
				},
			},
			ExpectedFormat: data.FormatListString,
			ExpectedData:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}

// TestDriftGenuineDriftContainsBothChangedLines exercises the exact diff
// content produced by a real Collect() run (rather than the harness's
// ElementsMatch-based ExpectedData, which is order-sensitive and awkward for
// diff text), confirming the unchanged placeholder line is masked out of the
// diff and only the genuinely-drifted line appears.
func TestDriftGenuineDriftContainsBothChangedLines(t *testing.T) {
	fact.Manager().ResetPlugins()
	defer fact.Manager().ResetPlugins()

	tmpl, _ := fact.Manager().GetPlugin("testdata:testfacter", "tmpl")
	tmplP := tmpl.(*testdata.TestFacter)
	tmplP.TestInputDataFormat = data.FormatRaw
	tmplP.TestInputData = []byte("name: {{ PROJECT_NAME }}\nruns-on: ubuntu-latest\n")
	tmplP.Collect()

	current, _ := fact.Manager().GetPlugin("testdata:testfacter", "current")
	currentP := current.(*testdata.TestFacter)
	currentP.TestInputDataFormat = data.FormatRaw
	currentP.TestInputData = []byte("name: my-actual-project\nruns-on: ubuntu-20.04\n")
	currentP.Collect()

	d := NewDrift("drift")
	d.SetInputName("tmpl")
	d.SetInput(tmplP)
	d.AdditionalInputNames = []string{"current"}
	errs := fact.LoadAdditionalInputs(d)
	assert.Empty(t, errs)

	d.Collect()
	assert.Empty(t, d.GetErrors())

	lines := data.AsListString(d.GetData())
	joined := ""
	var changedLines []string
	for _, l := range lines {
		joined += l
		if strings.HasPrefix(l, "+") || strings.HasPrefix(l, "-") {
			changedLines = append(changedLines, l)
		}
	}
	changedJoined := strings.Join(changedLines, "")
	assert.NotContains(t, changedJoined, "my-actual-project",
		"placeholder-masked line must appear only as unchanged diff context, not as a +/- change")
	assert.Contains(t, joined, "ubuntu-latest")
	assert.Contains(t, joined, "ubuntu-20.04")
}

func TestDriftUnsubstitutedPlaceholderMessage(t *testing.T) {
	fact.Manager().ResetPlugins()
	defer fact.Manager().ResetPlugins()

	tmpl, _ := fact.Manager().GetPlugin("testdata:testfacter", "tmpl")
	tmplP := tmpl.(*testdata.TestFacter)
	tmplP.TestInputDataFormat = data.FormatRaw
	tmplP.TestInputData = []byte("name: {{ PROJECT_NAME }}\n")
	tmplP.Collect()

	current, _ := fact.Manager().GetPlugin("testdata:testfacter", "current")
	currentP := current.(*testdata.TestFacter)
	currentP.TestInputDataFormat = data.FormatRaw
	currentP.TestInputData = []byte("name: {{ PROJECT_NAME }}\n")
	currentP.Collect()

	d := NewDrift("drift")
	d.SetInputName("tmpl")
	d.SetInput(tmplP)
	d.AdditionalInputNames = []string{"current"}
	errs := fact.LoadAdditionalInputs(d)
	assert.Empty(t, errs)

	d.Collect()
	assert.Empty(t, d.GetErrors())

	found := false
	for _, l := range data.AsListString(d.GetData()) {
		if l == `line 1: placeholder "{{ PROJECT_NAME }}" was left unsubstituted` {
			found = true
		}
	}
	assert.True(t, found, "expected an unsubstituted-placeholder notice, got %v", data.AsListString(d.GetData()))
}
