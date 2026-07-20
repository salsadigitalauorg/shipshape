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

These checks are already reproducible with existing plugins (status
**Achievable, undocumented** on the gaps matrix). They need **no new code** —
only a published recipe and a worked example.

Each shares the collect → parse → assert composition pattern shown in the
[`drupal-admin-user` recipe](gaps.md#recipe-drupal-admin-user-the-composition-pattern).

| 0.x check | Recipe to publish |
|---|---|
| `json` | `file:read` + `yaml:key` + `equals` / `allowed:list` |
| `drupal-db-module` | `command` + `allowed:list` |
| `drupal-db-permissions` | `command` / `database:search` + `allowed:list` |
| `drupal-db-user-tfa` | `command` + `equals` |
| `drupal-admin-user` | `command` + `yaml:key` + `allowed:list` |
| `drupal-user-forbidden` | `command` + `not:empty` / `equals` |
| `drupal-role-permissions` | `command` + `allowed:list` |
| `drupal-user-role` | `command` + `allowed:list` |
| `drupal-tracking-code` | `command` + `regex:match` / `not:empty` |

For each item, "done" means: a worked recipe in `examples/`, reference
documentation for the recipe, and the gaps matrix row flipped to
**Achievable now**.

## New general-purpose capabilities (deferred)

These checks have no existing plugin that can collect the data they need
(status **Needs new capability** on the gaps matrix). The intended direction is
to add a **general-purpose, reusable** collect plugin — not a check-specific
one — so that the new capability serves future recipes too. The exact approach
for each is **to be determined (TBD)**.

| 0.x check | Missing capability | Direction |
|---|---|---|
| `file:diff` | Compare a file against a rendered template and surface the difference | TBD |
| `crawler` | Crawl a site and collect non-200 responses | TBD |
| `sca:application_type` | Scan a codebase for framework markers and dependencies | TBD |

These are deferred with no committed timeline.

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