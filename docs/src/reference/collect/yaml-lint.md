# yaml:lint

The `yaml:lint` collect plugin reports which of its input files **fail to parse
as YAML**. It makes no assertion about the *contents* of those files — that is
[`yaml:key`](yaml-key.md)'s job. Use it as a cheap syntax gate over a directory
of config.

It is the 1.x replacement for the 0.x [`yamllint`](../checks/yaml-lint.md) check.

## Why this plugin exists

Everywhere else in the pipeline a YAML parse failure is a **collection error**,
and any collection error is fatal — the run aborts with `failed to collect facts`
before reaching the analyse stage. That makes the exact condition `yamllint`
reports impossible to express by composing other plugins.

`yaml:lint` inverts that relationship: a parse failure becomes ordinary **data**,
so an analyser can act on it. This is why the plugin never raises a collection
error for a malformed document — the only error it raises is genuine
misconfiguration (an unusable input format), which should still be fatal.

## Inputs

`yaml:lint` accepts a single input in `map-bytes` format — a map of filename to
file contents.

| Accepted format | Typical source |
| --- | --- |
| `map-bytes` | [`file:lookup`](file-lookup.md) with `file-names-only: false` |

File selection is deliberately delegated to `file:lookup`, so `path`, `pattern`,
`exclude-pattern` and `skip-dirs` are inherited rather than reimplemented.

::: warning `file-names-only` must be false
`file:lookup` defaults to emitting file *names* (`list-string`). `yaml:lint`
needs file *contents*, so set `file-names-only: false`. Leaving it at the
default is rejected at input validation with
`inputFormat 'list-string' not supported for 'yaml:lint'`.
:::

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| all-documents | Validate every document in a multi-document (`---`-separated) file rather than only the first. | No | `true` |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Return format

`map-string` — a map of **filename to parse error**, containing only the files
that failed. A tree with no invalid files emits an empty map (not nil), so
[`not:empty`](../analyse/not-empty.md) sees a supported format and simply finds
nothing to report.

Pair it with `not:empty`, which breaches once per entry.

## Example

```yaml
collect:
  config-files:
    file:lookup:
      path: config/default
      pattern: '.*\.yml$'
      file-names-only: false

  invalid-yaml:
    yaml:lint:
      input: config-files

analyse:
  yaml-is-valid:
    not:empty:
      description: 'YAML files that do not parse'
      input: invalid-yaml
      severity: high
      breach-format:
        type: key-value
        key-label: file
        key: '{{ .Breach.Key }}'
        value-label: 'YAML error'
        value: '{{ .Breach.Value }}'
```

Source: `examples/yaml-lint.yml` — see also the
[`yamllint` recipe](../../guide/gaps.md#recipe-yamllint).

## Error classification

Two message shapes are distinguished, preserving the 0.x breach labels:

| Condition | Message |
| --- | --- |
| A `yaml.TypeError` — the document parsed but contained conflicts, e.g. a duplicate key | `cannot decode yaml: <details>` |
| Any other parse failure — typically a syntax error | The parser message verbatim, e.g. `yaml: line 3: did not find expected ',' or ']'` |

A `TypeError` can aggregate several complaints; all are joined with `; ` rather
than discarding any.

::: tip Duplicate keys are reported
`a: 1` followed by `a: 2` is a `TypeError`, not silently accepted. This is worth
knowing because a duplicate key silently overrides the earlier value at runtime
— a real defect that is easy to miss in review.
:::

## Multi-document files

`all-documents` defaults to `true`, which is a **deliberate improvement on 0.x**.
The 0.x check used a single `yaml.Unmarshal`, which stops after the first
document — so a file whose *later* documents were malformed passed silently.
That matters for Kubernetes manifests and any `---`-separated config.

Set `all-documents: false` for bug-for-bug 0.x parity.

::: warning Reported line numbers can be one line early
Messages come verbatim from `gopkg.in/yaml.v3`, which for an unterminated
construct reports the line where that construct **opened**, not where parsing
gave up. An unclosed `[` on line 4 is reported as line 3. This is the library's
own behaviour — identical between `Unmarshal` and the decoder — and is passed
through unaltered rather than second-guessed. The 0.x check reported the same
positions.
:::

## Errors

| Condition | Behaviour |
| --- | --- |
| A file does not parse | **Not an error** — reported as data, which is the point of the plugin |
| Input format is not `map-bytes` | Rejected at input validation, before collection |
| Input path does not exist | A `file:lookup` collection error (fatal) — there is no `ignore-missing` equivalent |
| `pattern` matches zero files | **Passes.** `file:lookup` emits an empty `map-bytes`, so `yaml:lint` emits an empty `map-string`, so `not:empty` finds nothing to breach on |

::: warning A typo'd pattern reports green, not "nothing was checked"
This is a regression from 0.x, which explicitly breached when `pattern` matched
no files (`no matching yaml files found`). In 1.x, an existing path with a
`pattern` that matches nothing is indistinguishable from a directory that is
genuinely all valid YAML — both produce a passing `not:empty` result. A typo in
`pattern`, or a directory that has been restructured so the pattern no longer
matches anything, silently audits zero files rather than surfacing that as a
problem.

Only the *zero-match* case is affected. A `path` that does not exist at all is
still a `file:lookup` collection error and remains fatal (see the row above).

If validating that something was checked matters for your use case, assert on
the underlying `file:lookup` fact directly. There is no generic "list must be
non-empty" analyser, but if the expected filenames are known in advance,
`allowed:list` with a `required:` list against `file:lookup`'s `list-string`
output (`file-names-only` left at its default `true`) breaches
`required value not found` for any filename absent from the match — including
when nothing matched at all.
:::

## Differences from 0.x `yamllint`

| Aspect | 0.x | 1.x |
| --- | --- | --- |
| Multi-document files | First document only | All documents by default |
| Per-file passes | Adds `<file> has valid yaml.` per valid file | Only failures are reported; a clean run is a single pass on the analyser |
| File selection | `path`/`file`/`files`/`pattern`, plus `ignore-missing` | Delegated to `file:lookup`; `pattern` is required and a missing path is fatal |
| Zero files matched | Breaches — `no matching yaml files found` | **Passes** — see the warning above |

## Data classification

Breach output carries **file paths and parser messages**, and a parser message
can quote document content — a duplicate key's name, for example. If the linted
files hold sensitive configuration, any `output:` sink inherits that
classification. Treat accordingly where results are written to a lower
classification store than the audited files.
