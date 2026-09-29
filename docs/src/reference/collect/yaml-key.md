# yaml:key

The `yaml:key` collect plugin looks up a path in YAML data using a
[yaml-jsonpath](https://github.com/vmware-labs/yaml-jsonpath) expression and
emits the matched value(s). It is the YAML counterpart to
[`json:key`](json-key.md), using an older, pre-RFC dotted-path/filter dialect
rather than full RFC 9535 JSONPath — see the dialect warning below.

`yaml:key` is unusual in accepting **four** different input shapes, each
sourced from a different upstream fact — see [Input formats](#input-formats).

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| path | The yaml-jsonpath expression to evaluate against the input. | Yes | "" |
| nodes-only | Return the matched YAML nodes themselves rather than processed data. Used to feed a second `yaml:key` lookup (see [Chaining lookups](#chaining-lookups)). | No | false |
| keys-only | Return only the keys of the matched node, which must be a single mapping node. | No | false |
| ignore-not-found | Treat a not-found path as `nil` output rather than a collection error. | No | false |
| resolve-env | Resolve `${VAR}` references in matched scalar values from a `.env` file. | No | false |
| env-file | Path to the `.env` file used when `resolve-env` is true. Defaults to `<project-dir>/.env`. | No | "" |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Input formats

| Input format | Source | Behaviour |
| --- | --- | --- |
| `raw` | [`file:read`](file-read.md), [`http:fetch`](http-fetch.md) | A single YAML document; `path` is looked up directly. |
| `map-bytes` | [`file:lookup`](file-lookup.md) with `file-names-only: false` | Multiple YAML documents keyed by filename; `path` is looked up in each, producing a nested map keyed by filename. |
| `yaml-nodes` | A previous `yaml:key` call with `nodes-only: true` | A list of already-parsed YAML nodes (each expected to be a mapping node); `path` is looked up within each. |
| `map-yaml-nodes` | A previous `yaml:key` call with `nodes-only: true` against a `map-bytes` input | A map of lists of YAML nodes, keyed by filename; `path` is looked up within each file's nodes. |

## Return format

The emitted format is derived from the **shape of the match**:

| Match | Emitted format |
| --- | --- |
| A single scalar | `string` |
| A list of scalars | `list-string` |
| A list of mappings | `list-map-string` |
| A mapping | `map-string` |

When the input is `map-bytes` or `map-yaml-nodes`, the per-file result is
promoted one level — e.g. a scalar match per file becomes `map-string` (file
→ value) rather than `string`, and a mapping match per file becomes
`map-nested-string` (file → key → value).

`keys-only` always emits `list-string` — the mapping's keys — regardless of
input format, and only supports a single mapping node (`raw`/`yaml-nodes`
input); it errors against `map-bytes`/`map-yaml-nodes`.

`nodes-only` emits `yaml-nodes` (from `raw` input) or `map-yaml-nodes` (from
`map-bytes` input) — the raw matched nodes, unprocessed — for feeding into a
second `yaml:key` lookup.

## Chaining lookups

`nodes-only` lets one `yaml:key` fact narrow the document, and a second
`yaml:key` fact query within that narrowed scope — useful when a path needs
to be looked up per-entry across a YAML mapping or sequence, e.g. extracting
`image` from every service in a Docker Compose file:

```yaml
collect:
  compose-file:
    file:read:
      path: docker-compose.yml

  compose-services-nodes:
    yaml:key:
      input: compose-file
      path: services
      nodes-only: true

  images:
    yaml:key:
      input: compose-services-nodes
      path: image
```

Source: `examples/docker.yml`

## Dialect differs from `json:key`

`yaml:key` uses [yaml-jsonpath](https://github.com/vmware-labs/yaml-jsonpath),
an older pre-RFC JSONPath dialect. It accepts only the parenthesised filter
form (`[?(@.foo=='bar')]`) and supports none of RFC 9535's functions
(`length()`, `count()`, `match()`, `search()`). An expression written for
[`json:key`](json-key.md) will not necessarily work on `yaml:key` and vice
versa.

## `resolve-env` reads a `.env` file, not the OS environment

`resolve-env` (`pkg/env/envresolver.go`) reads **only** a `.env` file in the
project directory (or the path given by `env-file`) via `godotenv.Read`. It
never consults the shipshape process's own OS environment (`os.Environ()`).
A variable exported in the shell that runs `shipshape run` will **not** be
substituted into a matched value — it must be written to a `.env` file
instead.

## Example

```yaml
collect:
  test-file:
    file:read:
      path: pkg/fact/yaml/testdata/yaml-lookup.yml

  scalar-nodes:
    yaml:key:
      input: test-file
      path: scalar
      nodes-only: true

  list-string-nodes:
    yaml:key:
      input: test-file
      path: list-string
      nodes-only: true
```

Source: `examples/yaml-lookup.yml`

A second worked example, extracting the Lagoon service type from each
service in a Docker Compose file:

```yaml
collect:
  compose-file:
    file:read:
      path: docker-compose.yml

  compose-services-nodes:
    yaml:key:
      input: compose-file
      path: services
      nodes-only: true

  lagoon-type:
    yaml:key:
      input: compose-services-nodes
      path: labels['lagoon.type']

analyse:
  disallowed-lagoon-service:
    allowed:list:
      description: Disallowed Lagoon service type found in Docker Compose file.
      input: lagoon-type
      allowed:
        - none
        - cli
        - nginx
        - nginx-php
        - redis
        - solr
        - varnish
```

Source: `examples/docker.yml`

## Errors

| Condition | Behaviour |
| --- | --- |
| `path` not found | Collection error — `yaml path not found` — unless `ignore-not-found` is set, in which case the format is `nil`. |
| `keys-only` against a non-mapping node | Collection error — `keys-only lookup only supports a single mapping node`. |
| `keys-only` against `map-yaml-nodes` input | Collection error — `yaml-nodes-map unsupported format for keys-only lookup`. |
| A nested (`map-yaml-nodes`) lookup mixes incompatible per-file formats | Collection error — `unsupported format <format> for nested lookup`. |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage. Use
`ignore-not-found` when an absent path is an expected, non-fatal outcome.
:::