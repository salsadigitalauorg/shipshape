# file:read

The `file:read` collect plugin reads a single file from disk and emits its
content as raw bytes. It is a leaf collector (no `input`) — the most common
starting point for a pipeline, feeding [`yaml:key`](yaml-key.md),
[`json:key`](json-key.md), or [`file:drift`](file-drift.md).

For reading several known files at once, see
[`file:read:multiple`](file-read-multiple.md). For discovering files by
pattern rather than an exact path, see [`file:lookup`](file-lookup.md).

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| path | The file path to read, relative to the project directory. | Yes | "" |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Return format

`raw` — the file content as bytes.

## Path resolution

`path` is resolved relative to the project directory (the `-p`/`--path` flag
passed to `shipshape run`, defaulting to the current working directory), not
the current working directory of the shipshape process itself.

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
```

Source: `examples/yaml-lookup.yml`

## Errors

| Condition | Behaviour |
| --- | --- |
| File does not exist | Collection error wrapping `os.ErrNotExist`. |
| File exists but cannot be read (e.g. permissions) | Collection error wrapping the underlying `os.ReadFile` error. |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage. A
missing file therefore stops the run rather than producing a breach.
:::