# file:drift

The `file:drift` collect plugin compares a **template** against a **current
file** and emits the lines that differ, as a unified diff. Placeholder-shaped
substrings are masked out of both sides before comparing, so the same config runs
unmodified across every project provisioned from a shared template — no
per-project variables map required.

It is the 1.x replacement for the 0.x [`filediff`](../checks/file-diff.md) check.
Unlike `filediff` it does **not** render the template: see
[why rendering was rejected](../../guide/gaps.md#recipe-file-diff).

## Inputs

`file:drift` is unusual in taking **two** inputs:

| Input | Role | Accepted formats |
| --- | --- | --- |
| `input` | The template (the known-good baseline) | `raw`, `string` |
| `additional-inputs` | The current file being checked — exactly one | `raw`, `string` |

Both typically come from [`file:read`](file-read.md) or
[`http:fetch`](http-fetch.md).

::: warning Exactly one additional input, and its format is checked
Supplying zero or more than one `additional-inputs` is a collection error.
Pointing it at a multi-file fact (e.g. `file:lookup`, which emits `map-bytes`)
is also rejected rather than silently producing an empty "current" side — which
would otherwise breach claiming the whole template had been deleted, the worst
failure mode for an audit tool.
:::

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| placeholder-pattern | Regex matching placeholders to mask before diffing. | No | <code v-pre>{{\s*(\w+)\s*}}</code> |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Return format

`list-string` — the unified diff lines (3 lines of context), empty when there is
no drift. Pair it with the [`drift`](../analyse/drift.md) analyser, which
breaches whenever the list is non-empty.

## Example

```yaml
collect:
  template:
    http:fetch:
      url: https://raw.githubusercontent.com/client/project-template/main/.github/workflows/ci.yml

  current:
    file:read:
      path: .github/workflows/ci.yml

  ci-drift:
    file:drift:
      input: template
      additional-inputs: [current]

analyse:
  ci-matches-template:
    drift:
      description: CI workflow has not drifted from the project template
      input: ci-drift
```

Source: `examples/file-drift.yml`

## Placeholder matching is a wildcard, not a value check

Each placeholder is matched against the current line with a non-greedy `(.*?)`
capture. `file:drift` therefore confirms that *something* occupies the
placeholder's position — not that the rest of the line is otherwise identical in
structure. Consequences worth knowing:

- **A capture can absorb adjacent genuine drift on the same line.** Template
  <code v-pre>image: {{ IMG }}:v1</code> against current
  `image: evil/malware:latest:v1` reports no drift, because the wildcard
  swallows everything up to the literal `:v1`. Keep placeholders on their own
  line, or as an entire value, wherever the drift you care about could
  plausibly appear.
- **A repeated placeholder is not checked for consistency.** RE2 has no
  backreferences, so <code v-pre>a: {{ X }} b: {{ X }}</code> cannot express
  "these two captures must match" — current `a: one b: two` reports no drift.
- **An empty substitution counts as substituted**, not as drift. Only a
  placeholder still *literally present* in the current file — the
  copy-paste-and-forgot-to-fill-in bug — is flagged as unsubstituted.

These are inherent to wildcard matching without a rendering step, and are the
accepted trade-off against a per-project variables map.

## Output rendering

Drift breaches carry a full unified diff via `KeyValuesBreach`, legible in the
`pretty`, `json` and `junit` renderers. The `table` renderer does **not** wrap
multi-line breach values, so it is unsuitable for drift checks.