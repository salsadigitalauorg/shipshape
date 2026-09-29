# file:lookup

The `file:lookup` collect plugin walks a directory tree and finds files
matching a regular expression, optionally reading their content. It is a
leaf collector (no `input`) — used for discovering files whose exact paths
aren't known up front, e.g. finding disallowed scripts or sensitive files
by name pattern.

For reading a statically-known set of file paths, see
[`file:read`](file-read.md) or [`file:read:multiple`](file-read-multiple.md).

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| path | The directory to search, relative to the project directory. | Yes | "" |
| pattern | A regular expression matched against each file's basename. | No | "" |
| exclude-pattern | A regular expression matched against each file's basename; matching files are excluded even if `pattern` also matches. | No | "" |
| skip-dirs | Directory paths (relative to `path`) to skip entirely. | No | [] |
| file-names-only | Emit only the matched file paths rather than reading their content. | No | true |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Return format

| `file-names-only` | Format | Content |
| --- | --- | --- |
| `true` (default) | `list-string` | The matched file paths. |
| `false` | `map-bytes` | Matched file path to file content, for piping into [`yaml:key`](yaml-key.md), [`json:key`](json-key.md), or [`docker:images`](docker-images.md). |

## Example

Finding disallowed PHP admin scripts and sensitive files under a web root:

```yaml
collect:
  disallowed-php-scripts:
    file:lookup:
      path: web
      pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'
  sensitive-public-files:
    file:lookup:
      path: web/sites/default/files
      pattern: '.*\.(sql|php|sh|py|bz2|gz|tar|tgz|zip)?$'
      exclude-pattern: '.*\.(css|js)\.gz?$'
      skip-dirs:
        - private

analyse:
  disallowed-php-scripts-found:
    not:empty:
      description: 'Disallowed php scripts found'
      input: disallowed-php-scripts
      severity: high
  sensitive-public-files-found:
    not:empty:
      description: 'Sensitive files found in public directory'
      input: sensitive-public-files
      severity: high
```

Source: `examples/files.yml`

## `skip-dirs` matches by suffix path, not by name alone

Each entry in `skip-dirs` is joined onto `path` and matched against the
candidate file's path using `filepath.Rel`: a file is skipped only if its
path relative to `path`/`skip-dirs-entry` does not start with `..` — i.e.
the file is genuinely nested under that directory. A bare directory name
like `skipped` matches any directory with that name anywhere under `path`,
while a multi-segment entry like `nested-01/nested-02/skipped` matches only
that specific nested location.

## Errors

| Condition | Behaviour |
| --- | --- |
| The directory walk itself fails (e.g. `path` does not exist) | Collection error from the underlying walk. |
| An individual matched file cannot be read (only when `file-names-only: false`) | That file's error is recorded via `AddErrors`; the remaining matched files are still processed. |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage.
:::