package output_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestTableDisplay(t *testing.T) {
	tt := []struct {
		name     string
		rl       result.ResultList
		expected string
	}{
		{
			name:     "noResult",
			rl:       result.NewResultList(false),
			expected: "No result available; ensure your shipshape.yml is configured correctly.\n",
		},
		{
			name: "topShape",
			rl: result.ResultList{
				Results: []result.Result{{Name: "a", Status: result.Pass}},
			},
			expected: "NAME   STATUS   PASSES   FAILS\n" +
				"a      Pass              \n",
		},
		{
			name: "allPass",
			rl: result.ResultList{
				Results: []result.Result{
					{Name: "a", Status: result.Pass},
					{Name: "b", Status: result.Pass},
					{Name: "c", Status: result.Pass},
				},
			},
			expected: "NAME   STATUS   PASSES   FAILS\n" +
				"a      Pass              \n" +
				"b      Pass              \n" +
				"c      Pass              \n",
		},
		{
			name: "mixedPassFail",
			rl: result.ResultList{
				Results: []result.Result{
					{
						Name:   "a",
						Status: result.Pass,
						Passes: []string{"Pass a", "Pass ab"},
					},
					{
						Name:   "b",
						Status: result.Pass,
						Passes: []string{"Pass b", "Pass bb", "Pass bc"},
					},
					{
						Name:   "c",
						Status: result.Fail,
						Breaches: []breach.Breach{
							&breach.ValueBreach{Value: "Fail c"},
							&breach.ValueBreach{Value: "Fail cb"},
						},
					},
					{
						Name:   "d",
						Status: result.Fail,
						Passes: []string{"Pass d", "Pass db"},
						Breaches: []breach.Breach{
							&breach.ValueBreach{Value: "Fail c"},
							&breach.ValueBreach{Value: "Fail cb"},
						},
					},
				}},
			expected: "NAME   STATUS   PASSES    FAILS\n" +
				"a      Pass     Pass a    \n" +
				"                Pass ab   \n" +
				"b      Pass     Pass b    \n" +
				"                Pass bb   \n" +
				"                Pass bc   \n" +
				"c      Fail               Fail c\n" +
				"                          Fail cb\n" +
				"d      Fail     Pass d    Fail c\n" +
				"                Pass db   Fail cb\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			var buf bytes.Buffer
			s := &Stdout{}
			s.Table(&tc.rl, &buf)
			assert.Equal(tc.expected, buf.String())
		})
	}
}

func TestPrettyDisplay(t *testing.T) {
	tt := []struct {
		name     string
		rl       result.ResultList
		expected string
	}{
		{
			name:     "noResult",
			rl:       result.NewResultList(false),
			expected: "No result available; ensure your shipshape.yml is configured correctly.\n",
		},
		{
			name: "topShape",
			rl: result.ResultList{
				Results: []result.Result{{Name: "a", Status: result.Pass}},
			},
			expected: "Ship is in top shape; no breach detected!\n",
		},
		{
			name: "breachesDetected",
			rl: result.ResultList{
				Results: []result.Result{{
					Name:   "b",
					Status: result.Fail,
					Breaches: []breach.Breach{
						&breach.ValueBreach{Value: "Fail b"},
					},
				}},
			},
			expected: "# Breaches were detected\n\n  ### b\n     -- Fail b\n\n",
		},
		{
			name: "topShapeRemediating",
			rl: result.ResultList{
				Results:              []result.Result{{Name: "a", Status: result.Pass}},
				RemediationPerformed: true,
			},
			expected: "Ship is in top shape; no breach detected!\n",
		},
		{
			name: "allBreachesRemediated",
			rl: result.ResultList{
				Results: []result.Result{{
					Name: "a",
					Breaches: []breach.Breach{
						&breach.ValueBreach{
							RemediationResult: breach.RemediationResult{
								Status:   breach.RemediationStatusSuccess,
								Messages: []string{"fixed 1"},
							},
						},
					}}},
				TotalBreaches:        1,
				RemediationPerformed: true,
				RemediationTotals:    map[string]uint32{"successful": 1},
			},
			expected: "Breaches were detected but were all fixed successfully!\n\n" +
				"  ### a\n     -- fixed 1\n\n",
		},
		{
			name: "someBreachesRemediated",
			rl: result.ResultList{
				Results: []result.Result{{
					Name: "a",
					Breaches: []breach.Breach{
						&breach.ValueBreach{
							Value: "Fail a",
							RemediationResult: breach.RemediationResult{
								Status:   breach.RemediationStatusSuccess,
								Messages: []string{"fixed 1"},
							},
						},
						&breach.ValueBreach{
							Value: "Fail b",
							RemediationResult: breach.RemediationResult{
								Status:   breach.RemediationStatusFailed,
								Messages: []string{"not fixed 1"},
							},
						},
					}}},
				TotalBreaches:        2,
				RemediationPerformed: true,
				RemediationTotals:    map[string]uint32{"successful": 1, "failed": 1},
			},
			expected: "Breaches were detected but not all of them could be " +
				"fixed as they are either not supported yet or there were errors " +
				"when trying to remediate.\n\n" +
				"# Remediations\n\n  ### a\n     -- fixed 1\n\n" +
				"# Non-remediated breaches\n\n  ### a\n     -- Fail b\n\n",
		},
		{
			name: "noBreachRemediated",
			rl: result.ResultList{
				Results: []result.Result{{
					Name: "a",
					Breaches: []breach.Breach{
						&breach.ValueBreach{
							RemediationResult: breach.RemediationResult{
								Status:   breach.RemediationStatusFailed,
								Messages: []string{"failed 1"},
							},
						},
					}}},
				TotalBreaches:        1,
				RemediationPerformed: true,
				RemediationTotals:    map[string]uint32{"failed": 1},
			},
			expected: "Breaches were detected but none of them could be " +
				"fixed as there were errors when trying to remediate.\n\n" +
				"# Non-remediated breaches\n\n" +
				"  ### a\n     -- \n\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			var buf bytes.Buffer
			s := &Stdout{}
			s.Pretty(&tc.rl, &buf)
			assert.Equal(tc.expected, buf.String())
		})
	}
}

func TestJUnit(t *testing.T) {
	tt := []struct {
		name     string
		rl       result.ResultList
		expected string
	}{
		{
			name: "noResult",
			rl:   result.NewResultList(false),
			expected: `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0"></testsuites>
`,
		},
		{
			name: "allPass",
			rl: result.ResultList{
				Policies: map[string][]string{"test-check": {"a"}},
				Results:  []result.Result{{Name: "a", Status: result.Pass}}},
			expected: `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="test-check" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
    </testsuite>
</testsuites>
`,
		},
		{
			name: "mixedPassFail",
			rl: result.ResultList{
				Policies: map[string][]string{"test-check": {"a", "b"}},
				Results: []result.Result{
					{Name: "a", Status: result.Pass},
					{
						Name:     "b",
						Status:   result.Fail,
						Breaches: []breach.Breach{&breach.ValueBreach{Value: "Fail b"}},
					},
				},
			},
			expected: `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="test-check" tests="0" errors="0">
        <testcase name="a" classname="a"></testcase>
        <testcase name="b" classname="b">
            <error message="Fail b"></error>
        </testcase>
    </testsuite>
</testsuites>
`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			var buf bytes.Buffer
			s := &Stdout{}
			s.JUnit(&tc.rl, &buf)
			assert.Equal(tc.expected, buf.String())
		})
	}
}
