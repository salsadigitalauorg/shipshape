---
title: Checks (0.x)
---
# Checks (0.x) reference

These pages document the check types available in the legacy 0.x `checks:`
format. The format is **fully supported** and continues to run without changes —
see [Config versions](../../guide/0.x.md) for the format itself.

For new configurations, prefer the 1.x pipeline. The
[recipe catalogue](../../guide/gaps.md) shows the `collect` + `analyse` chain
that reproduces each check below.

::: warning Unrecognised check keys are silently ignored
An unknown key under `checks:` is discarded without warning — the run reports
zero checks and still exits `0`, which is indistinguishable from a passing build.
Copy the **Check type** value from each page exactly. Note in particular that
[`filediff`](file-diff.md) has no colon.
:::

## File checks

- [`file`](file.md) — assert files matching a pattern are absent
- [`filediff`](file-diff.md) — assert a file matches a reference template

## YAML / JSON checks

- [`yaml`](yaml.md) — assert key-value pairs in YAML files
- [`yamllint`](yaml-lint.md) — assert files parse as valid YAML
- [`json`](json.md) — assert key-value pairs in JSON files

## Drupal checks

- [`drush-yaml`](drupal-drush-yaml.md) — assert Drupal config via Drush
- [`drupal-file-module`](drupal-file-module.md) — assert modules via filesystem
- [`drupal-db-module`](drupal-db-module.md) — assert modules via database
- [`drupal-admin-user`](drupal-admin-user.md) — assert UID 1 roles
- [`drupal-db-permissions`](drupal-db-permissions.md) — assert role permissions
- [`drupal-db-user-tfa`](drupal-db-user-tfa.md) — assert users have TFA
- [`drupal-user-forbidden`](drupal-user-forbidden.md) — assert an account absent
- [`drupal-role-permissions`](drupal-role-permissions.md) — assert role permissions
- [`drupal-user-role`](drupal-user-role.md) — assert role holders are allow-listed
- [`drupal-tracking-code`](drupal-tracking-code.md) — assert tracking code present

## Code quality checks

- [`phpstan`](phpstan.md) — run PHPStan and report issues
- [`sca:application_type`](sca-application-type.md) — detect application framework

## Infrastructure checks

- [`docker:base_image`](docker-base-image.md) — assert approved base images
- [`crawler`](crawler.md) — assert pages return HTTP 200
