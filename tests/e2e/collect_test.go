package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// runCollect runs `shipshape collect` against the given working directory,
// pointing -f at the example file under examples/.
func runCollect(t *testing.T, workdir, exampleFile string, extraArgs ...string) RunResult {
	t.Helper()

	args := []string{"collect", ".", "-f", examplePath(exampleFile)}
	args = append(args, extraArgs...)
	return runShipshape(t, workdir, args...)
}

// TestCollectPrintsFactIds asserts that `collect` prints the user-assigned
// fact ID (e.g. "config-files", "invalid-yaml"), not the plugin type (e.g.
// "file:lookup", "yaml:lint").
func TestCollectPrintsFactIds(t *testing.T) {
	t.Parallel()

	res := runCollect(t, stagedTestdata(t), "yaml-lint.yml")

	assert.True(t, res.stdoutContains("config-files:"),
		"stderr: %s", res.Stderr)
	assert.True(t, res.stdoutContains("invalid-yaml:"),
		"stderr: %s", res.Stderr)
	assert.False(t, res.stdoutContains("file:lookup:"),
		"must not print the plugin type")
	assert.False(t, res.stdoutContains("yaml:lint:"),
		"must not print the plugin type")
}

// TestCollectFactsFilter asserts that `-n <fact-id>` filters collection to
// only the named fact, and that the fact ID (not the plugin type) is the
// value `-n` matches against.
func TestCollectFactsFilter(t *testing.T) {
	t.Parallel()

	res := runCollect(t, stagedTestdata(t), "yaml-lint.yml", "-n", "invalid-yaml")

	assert.True(t, res.stdoutContains("invalid-yaml:"), "stderr: %s", res.Stderr)
	assert.False(t, res.stdoutContains("config-files:"),
		"only the named fact should be collected and printed")
}
