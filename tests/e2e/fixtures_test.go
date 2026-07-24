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
// These tests assert the FIXED behaviour of the bugs discovered during the
// e2e overhaul (see docs/plans/2026-07-24-e2e-discovered-bugs-fixes.md):
//
//   - files.yml: not:empty now handles the FormatListString / FormatMapBytes
//     inputs emitted by file:lookup (pkg/analyse/notempty.go), so disallowed
//     files breach as intended.
//   - webforms: regex:match now handles FormatMapString inputs and
//     lookupFactAsStringMap is panic-safe, so the tokenised to_mail check
//     detects the token and renders its breach template cleanly.
//   - webforms cc-mail / bcc-mail: a handler that omits cc_mail / bcc_mail
//     now yields an empty input format, which regex:match treats as a no-op
//     (no breach) instead of emitting a ValueBreach whose key-value breach
//     template failed on the absent `.Breach.Key` field.

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

	// not:empty now fires for a file:lookup input (FormatListString), so both
	// the disallowed PHP script and the sensitive public file are flagged.
	assert.Equal(t, uint32(2), rl.TotalBreaches,
		"files.yml should breach on adminer.php and backup.sql")

	scripts, ok := findResult(rl, "disallowed-php-scripts-found")
	require.True(t, ok, "disallowed-php-scripts-found result not present")
	assert.Equal(t, "Fail", scripts.Status, "adminer.php must be flagged")

	sensitive, ok := findResult(rl, "sensitive-public-files-found")
	require.True(t, ok, "sensitive-public-files-found result not present")
	assert.Equal(t, "Fail", sensitive.Status, "backup.sql must be flagged")
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
	// A handler that DOES set a tokenised cc_mail: the cc-mail check must
	// breach, and its key-value breach template must render cleanly.
	writeFile(t, filepath.Join(project, "webform.webform.enquiry.yml"),
		"uuid: 3\nlangcode: en\nstatus: open\nid: enquiry\ntitle: 'Enquiry form'\n"+
			"handlers:\n  email_cc:\n    id: email\n    settings:\n      cc_mail: '[current-user:mail]'\n")

	res := runShipshape(t, repoRoot,
		"run", project, "-f", examplePath("webforms-tokenised-email-handlers.yml"), "-o", "json")
	rl := res.DecodeJSON(t)

	// The example wires three regex:match checks over the discovered webform
	// handlers and completes without a fatal error.
	assert.Equal(t, uint32(3), rl.TotalChecks, "webforms defines three checks")
	assert.Contains(t, rl.CheckCountByType, "regex:match")

	// regex:match now handles the FormatMapString input produced by yaml:key,
	// and lookupFactAsStringMap no longer panics: the tokenised to_mail is
	// detected and its breach template renders a real message (not a
	// "template" / "unsupported input format" error).
	toMail, ok := findResult(rl, "wrong-to-mail")
	require.True(t, ok, "wrong-to-mail result not present\nstderr: %s", res.Stderr)
	assert.Equal(t, "Fail", toMail.Status, "[site:mail] token must be flagged")
	require.NotEmpty(t, toMail.Breaches, "wrong-to-mail should have a breach")
	b := toMail.Breaches[0]
	assert.Equal(t, "[site:mail]", b["value"], "matched token value")
	// The rendered value-label proves both the FormatMapString handling and
	// the lookupFactAsStringMap template function worked without error.
	label, _ := b["value-label"].(string)
	assert.NotContains(t, label, "unable to render breach template",
		"breach template must render cleanly")
	assert.NotContains(t, label, "unsupported input format",
		"map-string input must be handled by regex:match")
	assert.Contains(t, label, "has token", "breach template must render the handler message")

	// bcc-mail: no fixture sets bcc_mail, so the yaml:key input is empty and
	// regex:match is a no-op. Previously this produced a spurious ValueBreach
	// whose key-value template failed on the absent .Breach.Key field.
	bccMail, ok := findResult(rl, "wrong-bcc-mail")
	require.True(t, ok, "wrong-bcc-mail result not present\nstderr: %s", res.Stderr)
	assert.Equal(t, "Pass", bccMail.Status, "absent bcc_mail must not breach")
	assert.Empty(t, bccMail.Breaches, "absent optional data is not a breach")

	// cc-mail: the enquiry fixture sets a tokenised cc_mail, so this check
	// must breach — and the key-value breach template must render cleanly
	// rather than erroring on a missing field.
	ccMail, ok := findResult(rl, "wrong-cc-mail")
	require.True(t, ok, "wrong-cc-mail result not present\nstderr: %s", res.Stderr)
	assert.Equal(t, "Fail", ccMail.Status, "tokenised cc_mail must be flagged")
	require.NotEmpty(t, ccMail.Breaches, "wrong-cc-mail should have a breach")
	ccb := ccMail.Breaches[0]
	assert.Equal(t, "[current-user:mail]", ccb["value"], "matched cc token value")
	ccLabel, _ := ccb["value-label"].(string)
	assert.NotContains(t, ccLabel, "unable to render breach template",
		"cc breach template must render cleanly")
	assert.NotContains(t, ccLabel, "unsupported input format",
		"map-string cc input must be handled by regex:match")
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
