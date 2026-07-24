//go:build e2e_docker

// Package e2e's Docker-daemon-dependent tests. These reproduce the drush-backed
// scenarios from the retired venom suite and are gated behind the e2e_docker
// build tag because they:
//   - require a working Docker daemon that can build images, and
//   - install a full Drupal site (composer create-project + drush si), which
//     takes minutes and needs a MariaDB backend.
//
// Run with:
//
//	go test -tags e2e_docker ./tests/e2e/... -timeout 20m
//
// The default `go test ./tests/e2e/...` (no tag) skips this file entirely so
// the everyday suite stays fast.
package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// drushExamples are the examples that require a live drush/Drupal environment.
// Reproducing the retired venom suite, we assert they run to a valid,
// non-fatal result against an installed site (rather than the fatal exit 127
// they produce with no drush on PATH).
var drushExamples = []string{
	"drupal-admin-user.yml",
	"drupal-db-module.yml",
	"drupal-db-permissions.yml",
	"drupal-role-permissions.yml",
	"drupal-user-forbidden.yml",
	"drupal-user-role.yml",
}

// TestDrushExamplesInDrupalContainer builds the Drupal+drush image, installs a
// site against a MariaDB container, and runs each drush-backed example inside
// the container. This is the Go-native replacement for the venom Drupal suite.
func TestDrushExamplesInDrupalContainer(t *testing.T) {
	ctx := context.Background()

	// 1. Backing database.
	const (
		dbName = "drupal"
		dbUser = "drupal"
		dbPass = "drupal"
	)
	net := newNetwork(t, ctx)

	db, err := mysql.Run(ctx,
		"uselagoon/mariadb-10.6-drupal",
		mysql.WithDatabase(dbName),
		mysql.WithUsername(dbUser),
		mysql.WithPassword(dbPass),
		network.WithNetwork([]string{"mariadb"}, net),
		testcontainers.WithWaitStrategy(
			wait.ForLog("ready for connections").AsRegexp().WithOccurrence(1).
				WithStartupTimeout(3*time.Minute),
		),
	)
	require.NoError(t, err, "start mariadb")
	t.Cleanup(func() { terminate(t, db) })

	// 2. Drupal + drush application container (reuses the committed Dockerfile).
	drupal := startDrupalContainer(t, ctx, net)

	// 3. Install the site.
	execInContainer(t, ctx, drupal,
		"drush", "site:install", "--yes", "--extra=\"--skip-ssl\"",
		fmt.Sprintf("--db-url=mysql://%s:%s@mariadb:3306/%s", dbUser, dbPass, dbName))

	// 4. Run each drush example inside the container against the installed site.
	for _, ex := range drushExamples {
		ex := ex
		t.Run(ex, func(t *testing.T) {
			code, stdout, stderr := execInContainerOutput(t, ctx, drupal,
				"shipshape", "run", "/app",
				"-f", "/shipshape/examples/"+ex, "-o", "json")
			assert.NotEqualf(t, 127, code,
				"drush should be available; example %s exited 127\nstderr: %s", ex, stderr)
			assert.NotEmptyf(t, stdout, "example %s produced no JSON\nstderr: %s", ex, stderr)
		})
	}
}
