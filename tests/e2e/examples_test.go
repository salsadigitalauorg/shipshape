package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// selfContainedCase describes a self-contained example: one that runs to a
// deterministic result using only fixtures committed under examples/testdata,
// with no external services. Expected values were captured empirically against
// the real binary (see docs/plans/2026-07-24-e2e-testcontainers-overhaul.md).
type selfContainedCase struct {
	name string
	file string
	// wantChecks / wantBreaches assert the top-level aggregates.
	wantChecks   uint32
	wantBreaches uint32
	// wantSeverity asserts the breach-count-by-severity map (nil to skip).
	wantSeverity map[string]int
	// wantPolicies asserts the policies map (nil to skip).
	wantPolicies map[string][]string
	// checkResults asserts on individual named results.
	checkResults []wantResult
	// wantExitClean asserts the exit code when run WITHOUT -e (always 0 for a
	// successful run) and WITH -e (0 when no breaches, 2 when breaches found).
	// Derived from wantBreaches, so no field needed.
}

type wantResult struct {
	name        string
	status      string // "Pass" or "Fail"
	checkType   string
	breachCount int
}

func selfContainedCases() []selfContainedCase {
	return []selfContainedCase{
		{
			name:         "regex-match",
			file:         "regex-match.yml",
			wantChecks:   1,
			wantBreaches: 1,
			wantSeverity: map[string]int{"normal": 1},
			wantPolicies: map[string][]string{"regex:match": {"lagoon-logs-check"}},
			checkResults: []wantResult{
				{name: "lagoon-logs-check", status: "Fail", checkType: "regex:match", breachCount: 1},
			},
		},
		{
			name:         "regex-not-match",
			file:         "regex-not-match.yml",
			wantChecks:   1,
			wantBreaches: 1,
			wantSeverity: map[string]int{"normal": 1},
			wantPolicies: map[string][]string{"regex:not-match": {"lagoon-logs-check"}},
			checkResults: []wantResult{
				{name: "lagoon-logs-check", status: "Fail", checkType: "regex:not-match", breachCount: 1},
			},
		},
		{
			name:         "required-values",
			file:         "required-values.yml",
			wantChecks:   1,
			wantBreaches: 0,
			wantPolicies: map[string][]string{"allowed:list": {"lagoon-logs-check"}},
			checkResults: []wantResult{
				{name: "lagoon-logs-check", status: "Pass", checkType: "allowed:list"},
			},
		},
		{
			name:         "json-lookup",
			file:         "json-lookup.yml",
			wantChecks:   3,
			wantBreaches: 2,
			// wantPolicies is intentionally omitted: the policy IDs within a
			// plugin are collected via Go map iteration (see shipshape.go), so
			// their order is non-deterministic. checkResults asserts each check
			// by name, which is order-independent.
			//
			// pinned-dependencies breaches twice: the RFC 9535 filter
			// `$.dependencies[?search(@,'\^')]` matches both caret-ranged
			// runtime dependencies in testdata/package.json.
			checkResults: []wantResult{
				{name: "approved-app-name", status: "Pass", checkType: "not:equals"},
				{name: "approved-script-tooling", status: "Pass", checkType: "allowed:list"},
				{name: "pinned-dependencies", status: "Fail", checkType: "not:empty", breachCount: 2},
			},
		},
		{
			name:         "drupal-config",
			file:         "drupal-config.yml",
			wantChecks:   3,
			wantBreaches: 1,
			wantSeverity: map[string]int{"high": 1, "normal": 0},
			checkResults: []wantResult{
				{name: "admin-role-found", status: "Fail", checkType: "allowed:list", breachCount: 1},
			},
		},
		{
			name:         "drupal-db-user-tfa",
			file:         "drupal-db-user-tfa.yml",
			wantChecks:   1,
			wantBreaches: 1,
			wantSeverity: map[string]int{"normal": 1},
			wantPolicies: map[string][]string{"equals": {"tfa-status-check"}},
			checkResults: []wantResult{
				{name: "tfa-status-check", status: "Fail", checkType: "equals", breachCount: 1},
			},
		},
		{
			name:         "drupal-tracking-code",
			file:         "drupal-tracking-code.yml",
			wantChecks:   1,
			wantBreaches: 1,
			wantSeverity: map[string]int{"normal": 1},
			wantPolicies: map[string][]string{"equals": {"tracking-code-check"}},
			checkResults: []wantResult{
				{name: "tracking-code-check", status: "Fail", checkType: "equals", breachCount: 1},
			},
		},
		{
			name:         "file-drift",
			file:         "file-drift.yml",
			wantChecks:   3,
			wantBreaches: 2,
			wantSeverity: map[string]int{"normal": 2},
			// wantPolicies is intentionally omitted: all three checks share
			// the drift plugin, and the policy IDs within a plugin are
			// collected via Go map iteration (see shipshape.go), so their
			// order is non-deterministic. checkResults asserts each check by
			// name, which is order-independent.
			checkResults: []wantResult{
				{name: "ci-matches-template", status: "Pass", checkType: "drift"},
				{name: "ci-drifted-from-template", status: "Fail", checkType: "drift", breachCount: 1},
				{name: "ci-placeholder-unsubstituted", status: "Fail", checkType: "drift", breachCount: 1},
			},
		},
	}
}

// TestSelfContainedExamples runs every self-contained example against the real
// binary and asserts on the structured JSON output. These are fast (no
// containers) and run in parallel.
func TestSelfContainedExamples(t *testing.T) {
	t.Parallel()

	for _, tc := range selfContainedCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res := runExample(t, stagedTestdata(t), tc.file, "json")
			rl := res.DecodeJSON(t)

			assert.Equalf(t, tc.wantChecks, rl.TotalChecks,
				"total-checks for %s\nstderr: %s", tc.file, res.Stderr)
			assert.Equalf(t, tc.wantBreaches, rl.TotalBreaches,
				"total-breaches for %s\nstderr: %s", tc.file, res.Stderr)

			if tc.wantSeverity != nil {
				assert.Equal(t, tc.wantSeverity, rl.BreachCountBySeverity,
					"breach-count-by-severity for %s", tc.file)
			}
			if tc.wantPolicies != nil {
				assert.Equal(t, tc.wantPolicies, rl.Policies,
					"policies for %s", tc.file)
			}

			for _, wr := range tc.checkResults {
				got, ok := findResult(rl, wr.name)
				require.Truef(t, ok, "result %q not found in %s", wr.name, tc.file)
				assert.Equalf(t, wr.status, got.Status,
					"status of %q in %s", wr.name, tc.file)
				if wr.checkType != "" {
					assert.Equalf(t, wr.checkType, got.CheckType,
						"check-type of %q in %s", wr.name, tc.file)
				}
				assert.Lenf(t, got.Breaches, wr.breachCount,
					"breach count of %q in %s", wr.name, tc.file)
			}
		})
	}
}

// TestSelfContainedExitCodes verifies the -e/--error-code contract: a clean run
// exits 0, and a run with breaches exits 2 (the code shipshape uses for
// detected failures at or above the fail-severity threshold).
func TestSelfContainedExitCodes(t *testing.T) {
	t.Parallel()

	for _, tc := range selfContainedCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Without -e, exit is always 0 regardless of breaches.
			clean := runExample(t, stagedTestdata(t), tc.file, "json")
			assert.Equalf(t, 0, clean.ExitCode,
				"exit code without -e for %s\nstderr: %s", tc.file, clean.Stderr)

			// With -e, exit reflects whether high-severity breaches were found.
			// shipshape.Exit exits 1 when a breach at/above fail-severity is
			// found and -e is set (see pkg/shipshape/shipshape.go:206).
			withE := runExample(t, stagedTestdata(t), tc.file, "json", "-e")
			if hasHighSeverityBreach(tc) {
				assert.Equalf(t, 1, withE.ExitCode,
					"expected exit 1 with -e for breaching example %s", tc.file)
			} else {
				assert.Equalf(t, 0, withE.ExitCode,
					"expected exit 0 with -e for %s\nstderr: %s", tc.file, withE.Stderr)
			}
		})
	}
}

// hasHighSeverityBreach reports whether the example is expected to breach at the
// default fail-severity ("high"). The -e flag only trips the error exit code
// when a breach meets the fail-severity threshold, so normal-severity breaches
// do not cause a non-zero exit under the default configuration.
func hasHighSeverityBreach(tc selfContainedCase) bool {
	return tc.wantSeverity["high"] > 0
}
