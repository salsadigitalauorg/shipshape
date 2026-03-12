package output_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/remediation"
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
							RemediationResult: remediation.RemediationResult{
								Status:   remediation.RemediationStatusSuccess,
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
					Name:            "a",
					DetailedMessage: "Custom message for a",
					Breaches: []breach.Breach{
						&breach.ValueBreach{
							Value: "Fail a",
							RemediationResult: remediation.RemediationResult{
								Status:   remediation.RemediationStatusSuccess,
								Messages: []string{"fixed 1"},
							},
						},
						&breach.ValueBreach{
							Value: "Fail b",
							RemediationResult: remediation.RemediationResult{
								Status:   remediation.RemediationStatusFailed,
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
				"# Non-remediated breaches\n\n  ### a\n     -- Fail b\n     #- Custom message for a\n\n",
		},
		{
			name: "noBreachRemediated",
			rl: result.ResultList{
				Results: []result.Result{{
					Name:            "a",
					DetailedMessage: "Custom message for a",
					Breaches: []breach.Breach{
						&breach.ValueBreach{
							RemediationResult: remediation.RemediationResult{
								Status:   remediation.RemediationStatusFailed,
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
				"  ### a\n     -- \n     #- Custom message for a\n\n",
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
        <testcase name="a" classname="test-check"></testcase>
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
        <testcase name="a" classname="test-check"></testcase>
        <testcase name="b" classname="test-check">
            <error message="Fail b"></error>
        </testcase>
    </testsuite>
</testsuites>
`,
		},
		{
			name: "singlePluginAllPass",
			rl: result.ResultList{
				TotalChecks: 3,
				Policies: map[string][]string{
					"regex:match": {"check1", "check2", "check3"},
				},
				CheckCountByType: map[string]int{
					"regex:match": 3,
				},
				BreachCountByType: map[string]int{
					"regex:match": 0,
				},
				Results: []result.Result{
					{Name: "check1", Status: result.Pass, CheckType: "regex:match"},
					{Name: "check2", Status: result.Pass, CheckType: "regex:match"},
					{Name: "check3", Status: result.Pass, CheckType: "regex:match"},
				},
			},
			expected: `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="3" errors="0">
    <testsuite name="regex:match" tests="3" errors="0">
        <testcase name="check1" classname="regex:match"></testcase>
        <testcase name="check2" classname="regex:match"></testcase>
        <testcase name="check3" classname="regex:match"></testcase>
    </testsuite>
</testsuites>
`,
		},
		{
			name: "singlePluginWithFailures",
			rl: result.ResultList{
				TotalChecks:   3,
				TotalBreaches: 2,
				Policies: map[string][]string{
					"regex:not-match": {"check1", "check2", "check3"},
				},
				CheckCountByType: map[string]int{
					"regex:not-match": 3,
				},
				BreachCountByType: map[string]int{
					"regex:not-match": 2,
				},
				Results: []result.Result{
					{
						Name:      "check1",
						Status:    result.Fail,
						CheckType: "regex:not-match",
						Breaches: []breach.Breach{
							&breach.ValueBreach{Value: "check1 failed"},
						},
						DetailedMessage: "Custom message for regex:not-match check1 failed",
					},
					{
						Name:      "check2",
						Status:    result.Fail,
						CheckType: "regex:not-match",
						Breaches: []breach.Breach{
							&breach.ValueBreach{Value: "check2 failed"},
						},
					},
					{Name: "check3", Status: result.Pass, CheckType: "regex:not-match"},
				},
			},
			expected: `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="3" errors="2">
    <testsuite name="regex:not-match" tests="3" errors="2">
        <testcase name="check1" classname="regex:not-match">
            <error message="check1 failed">Custom message for regex:not-match check1 failed</error>
        </testcase>
        <testcase name="check2" classname="regex:not-match">
            <error message="check2 failed"></error>
        </testcase>
        <testcase name="check3" classname="regex:not-match"></testcase>
    </testsuite>
</testsuites>
`,
		},
		{
			name: "multiplePlugins",
			rl: result.ResultList{
				TotalChecks:   4,
				TotalBreaches: 2,
				Policies: map[string][]string{
					"regex:match":     {"check1", "check2"},
					"regex:not-match": {"check3", "check4"},
				},
				CheckCountByType: map[string]int{
					"regex:match":     2,
					"regex:not-match": 2,
				},
				BreachCountByType: map[string]int{
					"regex:match":     1,
					"regex:not-match": 1,
				},
				Results: []result.Result{
					{Name: "check1", Status: result.Pass, CheckType: "regex:match"},
					{
						Name:      "check2",
						Status:    result.Fail,
						CheckType: "regex:match",
						Breaches: []breach.Breach{
							&breach.ValueBreach{Value: "check2 failed"},
						},
					},
					{Name: "check3", Status: result.Pass, CheckType: "regex:not-match"},
					{
						Name:      "check4",
						Status:    result.Fail,
						CheckType: "regex:not-match",
						Breaches: []breach.Breach{
							&breach.ValueBreach{Value: "check4 failed"},
						},
					},
				},
			},
			expected: `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="4" errors="2">
    <testsuite name="regex:match" tests="2" errors="1">
        <testcase name="check1" classname="regex:match"></testcase>
        <testcase name="check2" classname="regex:match">
            <error message="check2 failed"></error>
        </testcase>
    </testsuite>
    <testsuite name="regex:not-match" tests="2" errors="1">
        <testcase name="check3" classname="regex:not-match"></testcase>
        <testcase name="check4" classname="regex:not-match">
            <error message="check4 failed"></error>
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
