# 0.x to 1.x recipe catalogue

1.x does not provide a dedicated plugin for each 0.x check. Instead, you
reproduce a 0.x check by **composing general-purpose `collect` and `analyse`
plugins**. This page is the recipe catalogue: for every 0.x check it shows the
plugin chain that achieves the same outcome in 1.x.

See [The 1.x way: compose, don't port](versions.md#the-1-x-way-compose-don-t-port)
for the reasoning behind this approach.

::: tip Reading this table
None of the statuses below mean "waiting for a one-to-one port". Almost every
0.x check is reproducible today by composing plugins that already exist.

- **Achievable now** — a documented recipe and a working example in `examples/`
  exist. Copy the example and adapt it.
- **Achievable, undocumented** — the same plugins reproduce the check, but no
  worked example is published yet. The plugin chain is listed so you can build
  it; a published recipe is [on the roadmap](roadmap.md).
- **Needs new capability** — no existing building block collects the required
  data. The 1.x direction is still to be determined (TBD).
:::

## File checks

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`file`](../reference/checks/file.md) | Achievable now | `file:lookup` + `not:empty` — see `examples/files.yml` |
| [`file:diff`](../reference/checks/file-diff.md) | Needs new capability | TBD |

### Recipe: file

```yaml
collect:
  disallowed-php-scripts:
    file:lookup:
      path: web
      pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'

analyse:
  disallowed-php-scripts-found:
    not:empty:
      description: Disallowed PHP scripts found
      input: disallowed-php-scripts
```

Source: `examples/files.yml`

## YAML / JSON checks

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`yaml`](../reference/checks/yaml.md) | Achievable now | `file:read` + `yaml:key` + `equals` / `allowed:list` — see `examples/drupal-config.yml` |
| [`json`](../reference/checks/json.md) | Achievable now | `file:read` + `json:key` + `equals` / `allowed:list` — see `examples/json-lookup.yml` |

### Recipe: yaml

```yaml
collect:
  extension-file:
    file:read:
      path: core.extension.yml

  modules:
    yaml:key:
      input: extension-file
      path: module
      keys-only: true

analyse:
  required-modules:
    allowed:list:
      description: Required modules are not enabled
      input: modules
      required:
        - seckit
        - lagoon_logs
```

Source: `examples/drupal-config.yml`

### Recipe: json

`json:key` is the JSON counterpart to `yaml:key`. It reads raw JSON (typically
from a `file:read`) and evaluates an [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535.html)
JSONPath expression against it — e.g. `$.name`, `$.items[0]`, the wildcard
`$.scripts.*`, or a filter such as `$.deps[?@.name=='x'].version`.

The emitted format follows the shape of the match, mirroring `yaml:key`: a
single scalar emits a string, multiple values emit a list, and objects emit a
map. An expression that matches nothing emits nil rather than erroring, so
`not-empty` can act on the absence of a value.

Note that `json:key` implements RFC 9535, whereas `yaml:key` uses an older
pre-RFC dialect. The practical difference is filter syntax: `json:key` accepts
both `[?@.x=='y']` (the RFC form) and `[?(@.x=='y')]`, and supports the RFC
functions `length()`, `count()`, `match()` and `search()`; `yaml:key` accepts
only the parenthesised form and none of the functions.

```yaml
collect:
  pkg-file:
    file:read:
      path: package.json

  script-commands:
    json:key:
      input: pkg-file
      expression: "$.scripts.*"

analyse:
  approved-script-tooling:
    allowed:list:
      description: A script uses an unapproved build tool
      input: script-commands
      allowed:
        - eslint .
        - vite build
        - vitest run
```

Source: `examples/json-lookup.yml`

## Drupal checks

Every Drush-based Drupal check follows the same pattern: run a `command` (or
`docker:command`) to fetch data, parse it with `yaml:key` where needed, then
assert with `allowed:list` or `equals`. See the
[worked recipe](#recipe-drupal-admin-user-the-composition-pattern) below.

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`drush-yaml`](../reference/checks/drupal-drush-yaml.md) | Achievable now | `docker:command` + `yaml:key` + `equals` — see `examples/drush-over-docker.yml` |
| [`drupal-file-module`](../reference/checks/drupal-file-module.md) | Achievable now | `file:read` + `yaml:key` + `allowed:list` — see `examples/drupal-config.yml` |
| [`drupal-db-module`](../reference/checks/drupal-db-module.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-db-module.yml` |
| [`drupal-db-permissions`](../reference/checks/drupal-db-permissions.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-db-permissions.yml` |
| [`drupal-db-user-tfa`](../reference/checks/drupal-db-user-tfa.md) | Achievable now | `command` + `equals` — see `examples/drupal-db-user-tfa.yml` |
| [`drupal-admin-user`](../reference/checks/drupal-admin-user.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-admin-user.yml` |
| [`drupal-user-forbidden`](../reference/checks/drupal-user-forbidden.md) | Achievable now | `command` + `equals` — see `examples/drupal-user-forbidden.yml` |
| [`drupal-role-permissions`](../reference/checks/drupal-role-permissions.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-role-permissions.yml` |
| [`drupal-user-role`](../reference/checks/drupal-user-role.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-user-role.yml` |
| [`drupal-tracking-code`](../reference/checks/drupal-tracking-code.md) | Achievable now | `command` + `equals` — see `examples/drupal-tracking-code.yml` |

### Recipe: drush-yaml via Docker

```yaml
connections:
  docker-cli:
    docker:exec:
      container: my-app

collect:
  tfa-config:
    docker:command:
      connection: docker-cli
      command: ["/app/vendor/bin/drush", "config:get", "tfa.settings"]

  tfa-status:
    yaml:key:
      input: tfa-config
      path: enabled

analyse:
  tfa-enabled:
    equals:
      description: TFA must be enabled
      input: tfa-status
      value: "true"
```

Source: `examples/drush-over-docker.yml`

### Recipe: the Drush composition pattern

Every Drush-based check above follows the same shape: run a `command` fact that
emits the data one item per line, then assert with `allowed:list` or `equals`.

```yaml
collect:
  # 1. Collect: run drush and emit the machine name of each super-admin role.
  admin-roles:
    command:
      cmd: bash
      args:
      - -c
      - |
        set -o pipefail
        drush role:list --format=json \
          | jq -r 'to_entries[] | select(.value.is_admin == true) | .key'

analyse:
  # 2. Assert: fail if any role outside the allowed set is present.
  admin-roles-check:
    allowed:list:
      description: Unexpected role carries the is_admin flag
      input: admin-roles
      # `command` output is a map; `key: stdout` selects stdout, which
      # `allowed:list` then splits into one entry per line.
      key: stdout
      allowed:
        - administrator
```

Source: `examples/drupal-admin-user.yml`

Swap the `command` and the `allowed:list` / `equals` assertion to reproduce any
of the other Drush-based checks — the shape stays the same.

## Code quality checks

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`phpstan`](../reference/checks/phpstan.md) | Achievable now | `static-analysis` + `static-analysis:breaches` — see `examples/phpstan.yml` |
| [`sca:application_type`](../reference/checks/sca-application-type.md) | Needs new capability | TBD |

### Recipe: phpstan

```yaml
collect:
  phpstan-check:
    static-analysis:
      tool: phpstan
      config: phpstan.neon
      paths: [web/modules/custom]

analyse:
  phpstan-issues:
    static-analysis:breaches:
      description: PHPStan found code quality issues
      input: phpstan-check
      max-issues: 0
```

Source: `examples/phpstan.yml`

## Infrastructure checks

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`docker:base_image`](../reference/checks/docker-base-image.md) | Achievable now | `docker:images` + `allowed:list` — see `examples/docker.yml` |
| [`crawler`](../reference/checks/crawler.md) | Needs new capability | TBD |

### Recipe: docker:base_image

```yaml
collect:
  dockerfile-paths:
    yaml:key:
      input: compose-services-nodes
      path: build.dockerfile

  dockerfiles:
    file:read:multiple:
      input: dockerfile-paths

  base-images:
    docker:images:
      input: dockerfiles
      no-tag: true

analyse:
  disallowed-base-image:
    allowed:list:
      description: Disallowed base image found in Dockerfiles
      input: base-images
      allowed:
        - uselagoon/php-8.2-fpm
        - uselagoon/nginx-drupal
      deprecated:
        - uselagoon/php-8.1-fpm
```

Source: `examples/docker.yml`

## Summary

| Status | Count |
|---|---|
| Achievable now | 14 |
| Achievable, undocumented | 0 |
| Needs new capability | 4 |

The plan for the remaining capabilities is on the [roadmap](roadmap.md) page.