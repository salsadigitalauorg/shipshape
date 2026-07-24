package e2e

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// updateGolden regenerates the golden files instead of asserting against them.
// Run: go test ./tests/e2e/ -run TestRenderers -update
var updateGolden = flag.Bool("update", false, "update renderer golden files")

// rendererFormats is the set of human/machine output renderers locked by golden
// files. JSON is deliberately excluded: it is asserted structurally elsewhere.
// Adding a new format is a one-line change here plus regenerating goldens with
// -update.
var rendererFormats = []string{"pretty", "table", "junit"}

// rendererCases picks a small, representative spread of examples: one with a
// single breach, one that fully passes, and one multi-check example with a
// mix. This locks the renderers without coupling to every example.
//
// formats overrides rendererFormats for a case. drupal-config is excluded from
// junit because the JUnit renderer iterates the per-suite testcase map without
// sorting, so the order of multiple testcases within one testsuite is not
// stable across processes (Go randomises map iteration). See the discovered
// issue in docs/plans/2026-07-24-e2e-testcontainers-overhaul.md. The pretty and
// table renderers sort their rows and are stable.
var rendererCases = []struct {
	name    string
	file    string
	formats []string // nil => use rendererFormats
}{
	{name: "regex-match", file: "regex-match.yml"},         // single breach
	{name: "required-values", file: "required-values.yml"}, // all pass
	{name: "drupal-config", file: "drupal-config.yml", // multi-check, mixed
		formats: []string{"pretty", "table"}},
}

// TestRenderers locks the pretty, table and junit renderers with golden files.
// The examples chosen produce deterministic output (no timestamps, durations or
// absolute paths); normaliseRendererOutput defensively strips any that appear so
// the goldens stay stable across machines and CI.
func TestRenderers(t *testing.T) {
	for _, tc := range rendererCases {
		tc := tc
		formats := tc.formats
		if formats == nil {
			formats = rendererFormats
		}
		for _, format := range formats {
			format := format
			t.Run(tc.name+"/"+format, func(t *testing.T) {
				res := runExample(t, stagedTestdata(t), tc.file, format)
				require.Equalf(t, 0, res.ExitCode,
					"renderer run should exit 0\nstderr: %s", res.Stderr)

				got := normaliseRendererOutput(res.Stdout)
				goldenPath := filepath.Join("testdata", "golden", tc.name+"."+format+".golden")

				if *updateGolden {
					require.NoError(t, os.MkdirAll(filepath.Dir(goldenPath), 0o755))
					require.NoError(t, os.WriteFile(goldenPath, got, 0o644))
					return
				}

				want, err := os.ReadFile(goldenPath)
				require.NoErrorf(t, err,
					"missing golden file %s; regenerate with: go test ./tests/e2e/ -run TestRenderers -update",
					goldenPath)
				require.Equalf(t, string(want), string(got),
					"%s output for %s does not match golden %s\n(regenerate with -update if the change is intended)",
					format, tc.file, goldenPath)
			})
		}
	}
}

var (
	// repoRootPattern matches occurrences of the absolute repo root in output,
	// which would otherwise make goldens machine-specific.
	tmpPathPattern = regexp.MustCompile(`/(?:tmp|var/folders|private)/[^\s"'<>]+`)
	// junitTimePattern strips any time="..." attribute the JUnit renderer might
	// emit in future, keeping goldens deterministic.
	junitTimePattern = regexp.MustCompile(`\s+time="[^"]*"`)
)

// normaliseRendererOutput removes non-deterministic fragments (absolute temp
// paths, JUnit timing attributes, and the repo root) so golden comparisons are
// stable across environments.
func normaliseRendererOutput(b []byte) []byte {
	out := b
	out = junitTimePattern.ReplaceAll(out, nil)
	out = tmpPathPattern.ReplaceAll(out, []byte("<PATH>"))
	if repoRoot != "" {
		out = regexp.MustCompile(regexp.QuoteMeta(repoRoot)).ReplaceAll(out, []byte("<REPO>"))
	}
	return out
}
