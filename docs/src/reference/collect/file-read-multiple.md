# file:read:multiple

The `file:read:multiple` collect plugin reads several files from disk and
emits their contents as a map of filename to raw bytes. It complements
[`file:read`](file-read.md), which reads a single, statically-known path.

It can read either a static, operator-supplied list of files or a
dynamically-computed one — e.g. the file paths extracted by a
[`yaml:key`](yaml-key.md) lookup, letting the set of files to read be
determined by config rather than hardcoded.

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| files | A static list of file paths to read, relative to the project directory. Ignored if `input` is set. | No | [] |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Input formats

`input` is optional, but at least one of `input` or `files` must be
provided.

| Input format | Behaviour |
| --- | --- |
| `map-string` | Each value is treated as a file path; the map's **keys** are preserved in the output, mapping each original key to that file's content rather than the path. |

When no `input` is set (or the input carries no data), `files` is used
directly and the output map is keyed by the file path itself.

## Return format

`map-bytes` — filename (or, for a `map-string` input, the input's original
key) to file content.

## Example

Reading Dockerfiles whose paths were extracted per Compose service via
`yaml:key`:

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

  dockerfile-paths:
    yaml:key:
      input: compose-services-nodes
      path: build.dockerfile

  dockerfiles:
    file:read:multiple:
      input: dockerfile-paths
```

Source: `examples/docker.yml`

## Errors

Unlike [`file:read`](file-read.md), a missing or unreadable individual file
does **not** abort the whole collection — it is recorded via `AddErrors` and
that file is skipped, so the remaining files in the list are still read and
returned.

| Condition | Behaviour |
| --- | --- |
| Neither `input` nor `files` provided | Collection error — `no files specified`. |
| An individual file does not exist or cannot be read | That file's error is recorded; collection continues with the remaining files. |

::: warning A collection error still aborts the run
Even though a single missing file does not stop the whole list from being
processed, any recorded error is still fatal for the pipeline once
`Collect` returns — the run does not reach the analyse stage while facts
have errors attached.
:::