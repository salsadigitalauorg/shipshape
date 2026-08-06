# drift

The `drift` analyser breaches whenever its input is non-empty. It is designed to
pair with the [`file:drift`](../collect/file-drift.md) fact: any drift content at
all is itself the assertion failure.

## Plugin fields

`drift` has no plugin-specific fields — there is nothing to configure beyond the
common ones below.

::: tip No inverse polarity
Unlike `allowed:list` or `regex:match`, `drift` has no toggle to invert the
assertion. Drift always means breach.
:::

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/analyse.html').key"/>

## Example

```yaml
analyse:
  ci-matches-template:
    drift:
      description: CI workflow has not drifted from the project template
      input: ci-drift
```

Source: `examples/file-drift.yml` — see also the
[`filediff` recipe](../../guide/gaps.md#recipe-file-diff).

## Breach output

Breaches carry the full unified diff via `KeyValuesBreach`, legible in the
`pretty`, `json` and `junit` renderers. The `table` renderer does **not** wrap
multi-line breach values, so avoid it for drift checks.