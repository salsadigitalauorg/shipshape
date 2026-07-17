# json

Asserts key-value pairs in JSON files on the filesystem. Supports the same
assertion model as the [`yaml`](yaml.md) check.

**Check type:** `json`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `path` | string | one of path/file/files/pattern | Directory to search in |
| `file` | string | one of path/file/files/pattern | Single file name |
| `files` | list | one of path/file/files/pattern | Explicit list of file names |
| `pattern` | string | one of path/file/files/pattern | Regex pattern to match file names |
| `exclude-pattern` | string | no | Regex pattern — matching files are skipped |
| `key-values` | list of [KeyValue](#keyvalue-fields) | yes | Assertions to evaluate |

## KeyValue fields

Each entry in `key-values` supports:

| Field | Type | Description |
|---|---|---|
| `key` | string | Dot-separated path to the JSON key |
| `value` | string | Expected value (exact string match) |
| `truthy` | bool | If `true`, assert the value is truthy rather than matching `value` |
| `is-list` | bool | If `true`, treat the value at `key` as a list |
| `optional` | bool | If `true`, skip the assertion when the key is absent |
| `disallowed-values` | list | Values that must not appear at `key` |
| `allowed-values` | list | Values that must appear at `key` |

## Example

```yaml
checks:
  json:
    - name: Composer platform PHP version
      file: composer.json
      key-values:
        - key: config.platform.php
          value: "8.2"
```

## Behaviour

Shipshape locates the target file(s) using the supplied path/file/files/pattern
fields. For each file, it parses the JSON and evaluates every `key-values`
entry. Any assertion that fails is reported as a breach.

## Remediation

This check does not support automatic remediation. Update the JSON files
manually.