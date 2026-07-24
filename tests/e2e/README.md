# End-to-end tests

Go-native end-to-end tests for shipshape. The suite builds the real `shipshape`
binary once (`TestMain`) and runs it as a subprocess against every example in
[`examples/`](../../examples), asserting on the structured JSON output. Examples
that need a live service use [Testcontainers for Go](https://golang.testcontainers.org/).

## Running

Fast suite (self-contained examples + MariaDB via Testcontainers):

```sh
go test ./tests/e2e/...
```

Requires a working Docker daemon for the MariaDB-backed test
(`TestDomainInDbTablesExample`). Everything else runs as a plain subprocess.
Use `-short` to skip the container-backed tests:

```sh
go test -short ./tests/e2e/...
```

Heavy drush/Drupal suite (opt-in, builds a Drupal image and installs a site):

```sh
go test -tags e2e_docker -timeout 25m ./tests/e2e/...
```

## Layout

| File | Purpose |
|------|---------|
| `main_test.go` | Builds the shipshape binary once; disables the Testcontainers reaper for Colima compatibility |
| `harness.go` | Subprocess runner, JSON decode helpers, fixture helpers |
| `examples_test.go` | Self-contained examples (file/yaml/regex) — structural JSON assertions |
| `fixtures_test.go` | Examples needing an on-disk fixture tree (`files`, `webforms`, `docker`) |
| `mariadb_test.go` | `domain-in-db-tables` via a Testcontainers MariaDB |
| `renderers_test.go` | Golden-file tests for the `pretty`, `table` and `junit` renderers |
| `docker_test.go` + `docker_helpers_test.go` | Build-tagged (`e2e_docker`) drush/Drupal parity suite |
| `drupal.dockerfile` | Drupal + drush + shipshape image for the `e2e_docker` suite |
| `testdata/golden/` | Renderer golden snapshots |

## Assertion strategy

- **Correctness**: run with `-o json`, decode into a test-local `ResultList`
  mirror (the real `result.ResultList` cannot be unmarshalled because
  `Breaches` is an interface), assert on `total-checks`, `total-breaches`,
  per-check status and severity.
- **Renderers**: golden files for `pretty`, `table`, `junit`. Regenerate with:

  ```sh
  go test ./tests/e2e/ -run TestRenderers -update
  ```

  Only examples with deterministic output are goldened. `drupal-config` is
  excluded from `junit` because the JUnit renderer does not sort testcases
  within a suite (Go map-iteration order is not stable across processes).

## Known example issues surfaced by this suite

See `docs/plans/2026-07-24-e2e-testcontainers-overhaul.md` for detail:

- `files.yml` cannot breach: `file:lookup` emits `FormatListString` but the
  `not:empty` analyser only handles `FormatMapNestedString`.
- `webforms-tokenised-email-handlers.yml` emits template-render errors as breach
  content.
- The `junit` renderer does not sort testcases within a testsuite.

These tests assert current behaviour and will visibly change when the
underlying issues are fixed.
