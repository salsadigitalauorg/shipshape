package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDomainInDbTablesExample exercises examples/domain-in-db-tables.yml, which
// uses the database:search fact against a MySQL/MariaDB connection and breaches
// (via not:empty) when any text column contains a value matching the search
// pattern. It spins up a real MariaDB via Testcontainers, seeds a table that
// both has the id-field the example selects (entity_id) and a text column
// containing a .example.com value, then asserts shipshape finds it.
func TestDomainInDbTablesExample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping container-backed test in -short mode")
	}

	ctx := context.Background()

	const (
		dbName = "drupal"
		dbUser = "drupal"
		dbPass = "drupal"
	)

	// MariaDB's startup log differs from MySQL's, so the mysql module's default
	// log-based wait strategy times out. Wait on the port plus the MariaDB
	// "ready for connections" line, which both MySQL and MariaDB emit.
	container, err := mysql.Run(ctx,
		"mariadb:10.5",
		mysql.WithDatabase(dbName),
		mysql.WithUsername(dbUser),
		mysql.WithPassword(dbPass),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForListeningPort("3306/tcp"),
				wait.ForLog("ready for connections").AsRegexp().WithOccurrence(1),
			).WithDeadline(3*time.Minute),
		),
	)
	require.NoError(t, err, "start mariadb container")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("terminate mariadb container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	mappedPort, err := container.MappedPort(ctx, "3306/tcp")
	require.NoError(t, err)
	port := mappedPort.Port()

	seedDomainData(t, ctx, host, port, dbUser, dbPass, dbName)

	// Derive a config from the example with the connection host/port pointed at
	// the container. Everything else (search string, id-field, analyser) is
	// preserved from the committed example.
	cfg := domainInDbConfig(host, port, dbUser, dbPass, dbName)
	cfgPath := filepath.Join(t.TempDir(), "domain-in-db-tables.yml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfg), 0o644))

	res := runShipshape(t, repoRoot, "run", ".", "-f", cfgPath, "-o", "json")
	rl := res.DecodeJSON(t)

	assert.Equal(t, uint32(1), rl.TotalChecks, "one not:empty check\nstderr: %s", res.Stderr)
	assert.Equal(t, uint32(1), rl.TotalBreaches,
		"expected a breach for the seeded .example.com value\nstderr: %s", res.Stderr)

	got, ok := findResult(rl, "domain-found-in-tables")
	require.True(t, ok, "result domain-found-in-tables not present")
	assert.Equal(t, "Fail", got.Status)
	require.NotEmpty(t, got.Breaches, "expected at least one breach detail")
}

// seedDomainData connects directly to the container and creates a table that
// database:search will discover: it has an entity_id column (the example's
// id-field) and a varchar column holding a matching .example.com value.
func seedDomainData(t *testing.T, ctx context.Context, host, port, user, pass, dbName string) {
	t.Helper()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbName)

	var db *sql.DB
	var err error
	// The mysql module waits for readiness, but guard against transient dials.
	for i := 0; i < 10; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			if pingErr := db.Ping(); pingErr == nil {
				break
			} else {
				err = pingErr
			}
		}
		time.Sleep(time.Second)
	}
	require.NoError(t, err, "connect to seeded mariadb")
	t.Cleanup(func() { _ = db.Close() })

	stmts := []string{
		`CREATE TABLE config (entity_id INT PRIMARY KEY, name VARCHAR(255), data VARCHAR(255))`,
		`INSERT INTO config (entity_id, name, data) VALUES
			(1, 'system.site', 'https://legit.gov.au'),
			(2, 'domain.record', 'https://tenant.example.com/path')`,
	}
	for _, s := range stmts {
		_, err := db.ExecContext(ctx, s)
		require.NoErrorf(t, err, "seed statement failed: %s", s)
	}
}

// domainInDbConfig returns the domain-in-db-tables policy with a connection
// block pointing at the given host/port. Kept in sync with
// examples/domain-in-db-tables.yml (connection block aside).
func domainInDbConfig(host, port, user, pass, dbName string) string {
	return fmt.Sprintf(`connections:
  drupal-db:
    mysql:
      host: %s
      port: "%s"
      user: %s
      password: %s
      database: %s

collect:
  domain-in-tables:
    database:search:
      connection: drupal-db
      search: "%%.example.com%%"
      id-field: entity_id

analyse:
  domain-found-in-tables:
    not:empty:
      description: Domain found in table
      input: domain-in-tables
      breach-format:
        type: key-value
        key-label: Table
        key: ' {{ .Breach.Key }}'
        value-label: '[Column: {{ .Breach.ValueLabel }}]'
        value: 'Entity IDs: {{ .Breach.Value }}'
`, host, port, user, pass, dbName)
}
