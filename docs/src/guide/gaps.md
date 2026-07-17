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
| [`json`](../reference/checks/json.md) | Achievable, undocumented | `file:read` + `yaml:key` (parses JSON) + `equals` / `allowed:list` |

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

## Drupal checks

Every Drush-based Drupal check follows the same pattern: run a `command` (or
`docker:command`) to fetch data, parse it with `yaml:key` where needed, then
assert with `allowed:list` or `equals`. See the
[worked recipe](#recipe-drupal-admin-user-the-composition-pattern) below.

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`drush-yaml`](../reference/checks/drupal-drush-yaml.md) | Achievable now | `docker:command` + `yaml:key` + `equals` — see `examples/drush-over-docker.yml` |
| [`drupal-file-module`](../reference/checks/drupal-file-module.md) | Achievable now | `file:read` + `yaml:key` + `allowed:list` — see `examples/drupal-config.yml` |
| [`drupal-db-module`](../reference/checks/drupal-db-module.md) | Achievable, undocumented | `command` + `allowed:list` |
| [`drupal-db-permissions`](../reference/checks/drupal-db-permissions.md) | Achievable, undocumented | `command` / `database:search` + `allowed:list` |
| [`drupal-db-user-tfa`](../reference/checks/drupal-db-user-tfa.md) | Achievable, undocumented | `command` + `equals` — pattern shown in `examples/remediation.yml` |
| [`drupal-admin-user`](../reference/checks/drupal-admin-user.md) | Achievable, undocumented | `command` + `yaml:key` + `allowed:list` |
| [`drupal-user-forbidden`](../reference/checks/drupal-user-forbidden.md) | Achievable, undocumented | `command` + `not:empty` / `equals` |
| [`drupal-role-permissions`](../reference/checks/drupal-role-permissions.md) | Achievable, undocumented | `command` + `allowed:list` |
| [`drupal-user-role`](../reference/checks/drupal-user-role.md) | Achievable, undocumented | `command` + `allowed:list` |
| [`drupal-tracking-code`](../reference/checks/drupal-tracking-code.md) | Achievable, undocumented | `command` + `regex:match` / `not:empty` |

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

### Recipe: drupal-admin-user — the composition pattern

This is the template for the "Achievable, undocumented" Drupal checks above.
The 0.x `drupal-admin-user` check asserts that UID 1 holds only permitted
roles. In 1.x you compose the same three building blocks — collect, parse,
assert — that every other Drush check uses:

```yaml
collect:
  # 1. Collect: run drush to fetch UID 1's account information as YAML.
  admin-user:
    command:
      cmd: drush
      args: ["user:information", "1", "--format=yaml"]

  # 2. Parse: extract the roles field from the command output.
  admin-user-roles:
    yaml:key:
      input: admin-user
      path: roles

analyse:
  # 3. Assert: fail if any role outside the allowed set is present.
  admin-user-roles-allowed:
    allowed:list:
      description: UID 1 has a disallowed role
      input: admin-user-roles
      allowed:
        - authenticated
```

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
| Achievable now | 6 |
| Achievable, undocumented | 9 |
| Needs new capability | 3 |

The plan for publishing the undocumented recipes and deciding the direction for
the remaining capabilities is on the [roadmap](roadmap.md) page.