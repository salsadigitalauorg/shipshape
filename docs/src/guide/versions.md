# Config versions

Shipshape has two configuration formats. The same binary runs both.

::: tip 1.x is the recommended format
The 1.x pipeline (`collect` / `analyse` / `output`) is used in production and
is the recommended format for new configurations. The 0.x `checks:` format is
fully supported and continues to run without changes.

Not sure if 1.x covers everything you need? Check the
[gaps matrix](gaps.md) to see how each 0.x check is reproduced in 1.x, and the
[roadmap](roadmap.md) to see what's coming.
:::

## Which format should I use?

- **Starting a new configuration** — use 1.x.
- **Running an existing 0.x config** — it still runs as-is. Migrate to 1.x
  when convenient; there is no forced migration.
- **Relying on a check that isn't ported yet** — check the
  [gaps matrix](gaps.md). If the 1.x equivalent isn't there yet, keep using
  the 0.x check until it lands.

## How the binary picks a format

The deciding key is **`collect:`**. When Shipshape reads a config file it tries
to parse it as a 1.x config first: if a **non-empty top-level `collect:` block**
is present, the file is treated as 1.x and the pipeline runner is used.
Otherwise the file is treated as 0.x and the checks runner is used.

::: warning `collect:` is the only key that selects 1.x
`connections:`, `analyse:` and `output:` do **not** on their own make a file
1.x. A file containing only `analyse:` (or only `output:`) falls through to the
0.x runner, which finds no `checks:` block and runs zero checks — exiting `0`
as though everything passed.

Those three keys matter only once a run is already 1.x, where they are used to
reject a 0.x file passed alongside a 1.x one (see
[Mixing formats](#mixing-formats-across-files) below). An empty `collect:` block
does not count either — it must contain at least one fact.
:::

The distinguishing key is the top-level block name:

```yaml
# 0.x — identified by the top-level "checks:" key
checks:
  file:
    - name: Illegal files
      path: web
      disallowed-pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'
```

```yaml
# 1.x — identified by a non-empty top-level "collect:" key
collect:
  disallowed-php-scripts:
    file:lookup:
      path: web
      pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'

analyse:
  disallowed-php-scripts-found:
    not:empty:
      description: 'Disallowed php scripts found'
      input: disallowed-php-scripts
      severity: high
```

If a single file contains both a `checks:` key and a non-empty `collect:` key,
the 1.x runner takes precedence and the `checks:` block is ignored.

## Mixing formats across files

Passing a 0.x file and a 1.x file to the same run with `-f` is **an error, not a
merge**:

```sh
$ shipshape run . -f pipeline.yml -f legacy-checks.yml
config file "legacy-checks.yml" is not v2-compatible but was provided
alongside a v2 config; mixing v1 and v2 config files is not supported
```

Once any file in the run declares a non-empty `collect:`, every other file must
contain at least one of `connections`, `collect`, `analyse` or `output`. A file
with none of them is reported rather than silently discarded. Run the two
formats as separate invocations instead.

## Format overview

| | 0.x | 1.x |
|---|---|---|
| Top-level key | `checks:` | `collect:` (the deciding key) |
| Config struct | `Config` | `ConfigV2` |
| Runner | `RunConfig` | `RunV2` |
| Model | One check per concern — collection, evaluation, and optional remediation bundled together | Composable pipeline — collect data with one plugin, analyse it with another |
| Status | Fully supported | Recommended |

## What's the difference?

A 0.x check is a single, opinionated unit: it knows how to collect the data it
needs, evaluate it, and (for some checks) fix problems it finds. A 1.x config
separates those steps — a `collect` plugin gathers raw data, an `analyse`
plugin evaluates it. This makes the pieces reusable and composable, at the cost
of needing more explicit wiring.

The [gaps matrix](gaps.md) shows which 0.x checks have a full 1.x equivalent
today, and the [0.x config guide](0.x.md) documents the `checks:` format in
full.

## The 1.x way: compose, don't port

0.x gave you one dedicated check type per task. 1.x takes a different approach:
a small set of general-purpose `collect` and `analyse` plugins that you wire
together to achieve the same outcome. You won't find a one-to-one replacement
plugin for each 0.x check — instead you compose the same handful of building
blocks (for example `command` + `allowed:list`, or `file:read` + `yaml:key` +
`equals`) in different ways. The building blocks are reusable across many
checks, so learning a few covers a lot of ground.

The [gaps matrix](gaps.md) is the recipe catalogue: for each 0.x check it shows
the plugin chain that reproduces it in 1.x.
