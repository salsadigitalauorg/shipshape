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

## Current status: one capability gap remains

18 of the 19 registered 0.x checks have a documented recipe and a working example
in `examples/`. The remaining gap is **`yamllint`**: it reports an undecodable
YAML file as a breach, but in 1.x a parse failure is a collect error, and any
collect error is fatal to the run (`pkg/shipshape/shipshape.go:171-173`) — so the
pipeline never reaches the analyse stage where the breach would be raised.

Closing it needs a way to treat a fact's collection error as analysable data
rather than a fatal condition: either an analyser that acts on a fact's error
state (`BaseAnalyser` already reads `p.input.GetErrors()`), or an opt-in
"tolerate collect errors" mode. See
[Recipe: yamllint](gaps.md#recipe-yamllint-not-yet-reproducible).

The capabilities that were previously deferred here have all landed:

| 0.x check | Capability added |
|---|---|
| `json` | `json:key` fact plugin (RFC 9535 JSONPath) |
| `filediff` | `file:drift` fact + `drift` analyser — placeholder-masking instead of template rendering |
| `crawler` | `http:crawl` fact plugin |
| `sca:application_type` | `file:fingerprint` fact + `detected` analyser |

The Drush-based Drupal checks similarly all have worked examples now. They share
the collect → assert composition pattern shown in the
[Drush composition recipe](gaps.md#recipe-the-drush-composition-pattern): a
`command` fact emits data one item per line, and `allowed:list` / `equals`
makes the assertion.

Remaining work is therefore **reference documentation quality**, not new
plugins — several `reference/collect/` and `reference/connection/` pages are
still title-only stubs, and a few plugins (`http:fetch`, `json:key`, `drift`,
`detected`) have no reference page at all.

## Known plugin limitations to investigate

These are defects found while validating the bundled examples against 1.x. They
are tracked here so the affected examples can be simplified once the underlying
plugin work lands.

::: danger Both are panics, not errors
Each item below crashes the process on operator-supplied config rather than
reporting a collection error. A malformed or merely unusual config file should
never panic — these are correctness bugs, not cosmetic ones.
:::

| Area | Symptom | Direction |
|---|---|---|
| `yaml:key` — empty sequence | A config file with an empty YAML list (e.g. `permissions: []`) panics with `index out of range`. The `SequenceNode` branch indexes `Content[0]` without a length check in **four** places: `YamlLookup.ProcessNodes` (`pkg/fact/yaml/yaml.go:106`, `:122`) and `AliasNodeToData` (`:238`, `:254`). | Guard every `SequenceNode` case against empty `Content`, and emit an empty list rather than panicking. All four sites need the guard, not just the one reached first. |
| `allowed:list` — multi-file input | `yaml:key` output over a `file:lookup` (map of file → list) panics in `AsMapListString` (`pkg/data/data.go:94`) with `interface conversion: map[string]interface{}, not map[string][]string`, via `allowedlist.go:178`. When any file has an empty map the format instead becomes `map-nested-string`, which `allowed:list` does not handle. | Make `allowed:list` accept the actual multi-file `yaml:key` data shape (and `map-nested-string`), so a "disallowed permission across all roles" assertion can be expressed without per-role single-file reads. This is why `examples/drupal-config.yml` asserts permissions per-role rather than across all roles at once. |

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