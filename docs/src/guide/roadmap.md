# Roadmap to 1.x parity

::: warning Direction, not a guarantee
This page describes the intended direction for reaching parity between the 0.x
`checks:` format and the 1.x pipeline. Scope and sequencing may change as
priorities shift and contributors engage.

- **1.x is already recommended for production.** See [Config versions](versions.md).
- **0.x is fully supported** while this work progresses. Nothing will break.
- The current recipe catalogue is on the [gaps matrix](gaps.md).
:::

## The direction: recipes, not plugin ports

1.x reaches parity by **documenting composition recipes**, not by building a
dedicated plugin for each 0.x check. Almost every 0.x check is already
reproducible today by wiring together general-purpose `collect` and `analyse`
plugins — see the [gaps matrix](gaps.md) for the plugin chain per check.

So most of the remaining work is **writing docs and examples**, not writing
code. Building a new plugin is the exception, reserved for the few checks where
no existing building block can collect the required data.

## What "parity" means

Parity is reached when every 0.x check has a documented, tested recipe composed
from general-purpose plugins — and, where the 0.x check supported it, a
[remediation](remediate.md) path. A recipe counts as complete when it is
documented in the reference and backed by a working file in `examples/`.

The [gaps matrix](gaps.md) is the authoritative record of current status.

## Recipes to document

The Drush-based Drupal checks that were previously **Achievable, undocumented**
now each have a worked example in `examples/` and are marked **Achievable now**
on the gaps matrix:

| 0.x check | Example |
|---|---|
| `drupal-db-module` | `examples/drupal-db-module.yml` |
| `drupal-db-permissions` | `examples/drupal-db-permissions.yml` |
| `drupal-db-user-tfa` | `examples/drupal-db-user-tfa.yml` |
| `drupal-admin-user` | `examples/drupal-admin-user.yml` |
| `drupal-user-forbidden` | `examples/drupal-user-forbidden.yml` |
| `drupal-role-permissions` | `examples/drupal-role-permissions.yml` |
| `drupal-user-role` | `examples/drupal-user-role.yml` |
| `drupal-tracking-code` | `examples/drupal-tracking-code.yml` |

They all share the collect → assert composition pattern shown in the
[Drush composition recipe](gaps.md#recipe-the-drush-composition-pattern): a
`command` fact emits data one item per line, and `allowed:list` / `equals`
makes the assertion.

## New general-purpose capabilities (deferred)

These checks have no existing plugin that can collect the data they need
(status **Needs new capability** on the gaps matrix). The intended direction is
to add a **general-purpose, reusable** collect plugin — not a check-specific
one — so that the new capability serves future recipes too. The exact approach
for each is **to be determined (TBD)**.

| 0.x check | Missing capability | Direction |
|---|---|---|
| `json` | Parse JSON files and extract keys/values | Add a general-purpose `json:key` fact plugin (analogous to `yaml:key`) |
| `file:diff` | Compare a file against a rendered template and surface the difference | TBD |
| `crawler` | Crawl a site and collect non-200 responses | TBD |
| `sca:application_type` | Scan a codebase for framework markers and dependencies | TBD |

These are deferred with no committed timeline.

## Known plugin limitations to investigate

These are defects and gaps found while validating the bundled examples against
1.x. They are tracked here so the affected examples can be simplified or
completed once the underlying plugin work lands.

| Area | Symptom | Direction |
|---|---|---|
| `examples/docker.yml` — `base-images` | `docker:images` with `additional-inputs: [buildargs]` fails at collect with `inputFormat required for 'yaml:key'`. | Investigate whether the example needs an explicit `input-format` on the additional input, or whether `docker:images` additional-input handling regressed. The `base-images` block is currently commented/omitted from a runnable path. |
| `yaml:key` — empty sequence | A role/config file with an empty YAML list (e.g. `permissions: []`) panics in `YamlLookup.ProcessNodes` (index out of range). | Guard the `SequenceNode` case against empty `Content`. |
| `allowed:list` — multi-file input | `yaml:key` output over a `file:lookup` (map of file → list) produces `map-nested-string` when any file has an empty map, which `allowed:list` does not support; when all files have lists, `allowed:list`'s `map-list-string` branch panics in `AsMapListString` (the fact data is `map[string]interface{}`, not `map[string][]string`). | Make `allowed:list` accept the actual multi-file `yaml:key` data shape (and `map-nested-string`), so a "disallowed permission across all roles" assertion can be expressed without per-role single-file reads. This is why `examples/drupal-config.yml` asserts permissions per-role rather than across all roles at once. |

## How to contribute

The [gaps matrix](gaps.md) is the source of truth for what remains open.

To pick up an item:

1. Check the [issue tracker](https://github.com/salsadigitalauorg/shipshape/issues)
   for an existing issue before opening a new one.
2. Reference the 0.x check type (e.g. `drupal-admin-user`) in the issue title
   so it stays traceable to this roadmap.

An item is **done** when:

- **Recipe items** — the recipe is documented in `docs/src/reference/`, a
  working file exists in `examples/`, and the gaps matrix row is updated to
  **Achievable now**.
- **Capability items** — a new **general-purpose** collect plugin is registered
  and documented (not a check-specific plugin), with a recipe showing how it
  reproduces the 0.x check.

Contributions to either group are welcome. There is no strict ordering — pick
whatever is most useful to you first.