# yamllint

Asserts that files parse as valid YAML. Unlike [`yaml`](yaml.md), it makes no
assertions about the *contents* — it only reports files that cannot be decoded.
Use it as a cheap syntax gate over a directory of config before other checks try
to read it.

**Check type:** `yamllint`

## Fields

`yamllint` accepts the same file-selection fields as [`yaml`](yaml.md), since it
embeds the same check:

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `path` | string | one of path/file/files/pattern | Directory to search in |
| `file` | string | one of path/file/files/pattern | Single file name |
| `files` | list | one of path/file/files/pattern | Explicit list of file names |
| `pattern` | string | one of path/file/files/pattern | Regex pattern to match file names |
| `exclude-pattern` | string | no | Regex pattern — matching files are skipped |
| `ignore-missing` | bool | no | If `true`, pass instead of failing when a file or path does not exist |

A `values:` list, if supplied, is ignored — this check does not evaluate
assertions.

## Example

```yaml
checks:
  yamllint:
    - name: Exported config is valid YAML
      path: config/default
      pattern: '.*\.yml$'
```

A single file:

```yaml
checks:
  yamllint:
    - name: CI workflow parses
      file: .github/workflows/ci.yml
```

## Behaviour

Each selected file is unmarshalled with `gopkg.in/yaml.v3` into a generic
`interface{}`. A file that decodes cleanly adds a pass
(`<file> has valid yaml.`); a file that fails adds a breach. Two breach shapes
are distinguished:

- **`cannot decode yaml: <file>`** — a `yaml.TypeError`, meaning the document
  parsed but contained type conflicts. All reported type errors are joined into
  the breach value.
- **`yaml error: <file>`** — any other unmarshal error, typically a syntax
  error. The underlying parser message is the breach value.

The check passes when no file produced a breach.

## Remediation

This check does not support automatic remediation. Fix the reported YAML
manually.
