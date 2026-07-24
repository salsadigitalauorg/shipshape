package e2e

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These examples exercise the file:lookup fact plugin and need a small
// on-disk fixture tree that does not exist under examples/testdata. Each test
// seeds a temp directory and passes it as the project directory argument.
//
// IMPORTANT (discovered during implementation, see
// docs/plans/2026-07-24-e2e-testcontainers-overhaul.md): files.yml and
// webforms-tokenised-email-handlers.yml do not currently function as their
// comments describe against the v2 collect/analyse engine:
//
//   - files.yml pairs a file:lookup fact (which emits FormatListString /
//     FormatMapBytes) with the not:empty analyser, which only handles
//     FormatMapNestedString (pkg/analyse/notempty.go). The disallowed-files
//     check therefore can never breach, even when disallowed files are present.
//   - webforms produces breaches whose content is template-rendering errors
//     ("unable to render breach template", "unsupported input format").
//
// Until those are fixed upstream, these tests assert the CURRENT, stable
// behaviour (the run completes, emits valid JSON, and the aggregate shape is
// what the engine actually produces) so they act as regression guards and will
// visibly change if the underlying bugs are fixed.

func TestFilesExample(t *testing.T) {
	t.Parallel()

	project := t.TempDir()
	// A file matching the disallowed-php pattern.
	writeFile(t, filepath.Join(project, "web", "adminer.php"), "<?php\n")
	// A sensitive file in the public files directory.
	writeFile(t, filepath.Join(project, "web", "sites", "default", "files", "backup.sql"), "dump\n")
	// A benign file that must not be flagged.
	writeFile(t, filepath.Join(project, "web", "index.php"), "<?php\n")

	res := runShipshape(t, repoRoot,
		"run", project, "-f", examplePath("files.yml"), "-o", "json")
	rl := res.DecodeJSON(t)

	// Two not:empty checks are defined and both run.
	assert.Equal(t, uint32(2), rl.TotalChecks, "files.yml defines two checks")
	assert.Contains(t, rl.CheckCountByType, "not:empty")

	// Current behaviour: not:empty cannot fire for a file:lookup input, so
	// there are zero breaches. If this assertion starts failing, the upstream
	// not:empty/file:lookup format mismatch has likely been fixed and this
	// test (and the manifest) should be updated to assert real breaches.
	assert.Equal(t, uint32(0), rl.TotalBreaches,
		"files.yml currently cannot breach; see notempty.go format handling")
}

func TestWebformsExample(t *testing.T) {
	t.Parallel()

	project := t.TempDir()
	writeFile(t, filepath.Join(project, "webform.webform.contact.yml"),
		"uuid: 1\nlangcode: en\nstatus: open\nid: contact\ntitle: 'Contact form'\n"+
			"handlers:\n  email_confirmation:\n    id: email\n    settings:\n      to_mail: '[site:mail]'\n")
	writeFile(t, filepath.Join(project, "webform.webform.feedback.yml"),
		"uuid: 2\nlangcode: en\nstatus: open\nid: feedback\ntitle: 'Feedback form'\n"+
			"handlers:\n  email_notification:\n    id: email\n    settings:\n      to_mail: 'admin@example.com'\n")

	res := runShipshape(t, repoRoot,
		"run", project, "-f", examplePath("webforms-tokenised-email-handlers.yml"), "-o", "json")
	rl := res.DecodeJSON(t)

	// Smoke-level guard: the example wires three regex:match checks over the
	// discovered webform handlers and completes without a fatal error.
	assert.Equal(t, uint32(3), rl.TotalChecks, "webforms defines three checks")
	assert.Contains(t, rl.CheckCountByType, "regex:match")
	assert.Equal(t, 0, res.ExitCode, "run completes without fatal error")
}

// dockerComposeFixture is a minimal Lagoon-style compose project plus the
// Dockerfiles it references. It contains two allowed services and one
// disallowed image (evil/malware:latest) to exercise the allowed:list check.
const dockerComposeFixture = `services:
  cli:
    image: uselagoon/php-8.1-cli-drupal
    build:
      dockerfile: cli.dockerfile
      args:
        CLI_IMAGE: uselagoon/php-8.1-cli-drupal
    labels:
      lagoon.type: cli-persistent
  nginx:
    image: uselagoon/nginx-drupal
    build:
      dockerfile: nginx.dockerfile
      args:
        CLI_IMAGE: uselagoon/php-8.1-cli-drupal
    labels:
      lagoon.type: nginx
  badservice:
    image: evil/malware:latest
    labels:
      lagoon.type: nginx
`

// TestDockerComposeExample exercises examples/docker.yml. Despite the name, the
// docker:images fact only parses FROM lines from Dockerfile *content* — it does
// not talk to the Docker daemon — so this runs as a self-contained fixture test.
func TestDockerComposeExample(t *testing.T) {
	t.Parallel()

	project := t.TempDir()
	writeFile(t, filepath.Join(project, "docker-compose.yml"), dockerComposeFixture)
	writeFile(t, filepath.Join(project, "cli.dockerfile"), "FROM uselagoon/php-8.1-cli-drupal\n")
	writeFile(t, filepath.Join(project, "nginx.dockerfile"), "FROM uselagoon/nginx-drupal\n")

	res := runShipshape(t, repoRoot,
		"run", project, "-f", examplePath("docker.yml"), "-o", "json")
	rl := res.DecodeJSON(t)

	assert.Equal(t, uint32(3), rl.TotalChecks,
		"docker.yml defines three checks\nstderr: %s", res.Stderr)

	// The disallowed image (evil/malware) must be flagged.
	got, ok := findResult(rl, "disallowed-service-image")
	require.True(t, ok, "disallowed-service-image result not present")
	assert.Equal(t, "Fail", got.Status, "evil/malware should be disallowed")

	// The two Lagoon-approved base images must pass their checks.
	if base, ok := findResult(rl, "disallowed-base-image"); ok {
		assert.Equal(t, "Pass", base.Status, "approved base images should pass")
	}
}
