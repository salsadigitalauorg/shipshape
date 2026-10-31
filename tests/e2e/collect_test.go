package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCollect runs `shipshape collect` against the given working directory,
// pointing -f at the example file under examples/.
func runCollect(t *testing.T, workdir, exampleFile string, extraArgs ...string) RunResult {
	t.Helper()

	args := []string{"collect", ".", "-f", examplePath(exampleFile)}
	args = append(args, extraArgs...)
	return runShipshape(t, workdir, args...)
}

// lagoonFactsPayload mirrors cmd.lagoonFactsPayload for decoding
// `collect -o lagoon-facts` output.
type lagoonFactsPayload struct {
	Project     string `json:"project"`
	Environment string `json:"environment"`
	Facts       []struct {
		Name        string `json:"name"`
		Value       string `json:"value"`
		Source      string `json:"source"`
		Category    string `json:"category"`
		Description string `json:"description"`
	} `json:"facts"`
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

// TestCollectLagoonFacts asserts that `collect -o lagoon-facts` emits valid
// JSON matching the addFactsByName mutation input shape, with no network call
// made (LAGOON_PROJECT/LAGOON_ENVIRONMENT is all that's required, no API
// credentials).
func TestCollectLagoonFacts(t *testing.T) {
	t.Parallel()

	res := runShipshapeEnv(t, stagedTestdata(t),
		[]string{"LAGOON_PROJECT=test-shipshape", "LAGOON_ENVIRONMENT=master"},
		"collect", ".", "-f", examplePath("yaml-lint.yml"), "-o", "lagoon-facts")

	require.Equal(t, 0, res.ExitCode, "stderr: %s", res.Stderr)

	var payload lagoonFactsPayload
	err := json.Unmarshal(res.Stdout, &payload)
	require.NoError(t, err, "stdout: %s\nstderr: %s", res.Stdout, res.Stderr)

	assert.Equal(t, "test-shipshape", payload.Project)
	assert.Equal(t, "master", payload.Environment)
	require.Len(t, payload.Facts, 3)
	for _, f := range payload.Facts {
		assert.Equal(t, "Shipshape", f.Source)
		assert.Equal(t, "invalid-yaml", f.Category)
		assert.NotEmpty(t, f.Name)
		assert.NotEmpty(t, f.Value)
	}
}

// TestCollectLagoonFactsFailsFastWithoutEnvVars asserts that `collect -o
// lagoon-facts` fails fast with a clear message when LAGOON_PROJECT/
// LAGOON_ENVIRONMENT are not set, rather than emitting a payload with empty
// project/environment that addFactsByName would reject.
func TestCollectLagoonFactsFailsFastWithoutEnvVars(t *testing.T) {
	t.Parallel()

	res := runCollect(t, stagedTestdata(t), "yaml-lint.yml", "-o", "lagoon-facts")

	assert.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, string(res.Stderr),
		"project & environment name required")
}
