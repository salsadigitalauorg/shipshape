# yaml

Asserts key-value pairs in YAML files on the filesystem. Use this to verify
Drupal exported config, application settings, or any structured YAML without
needing a running Drupal instance.

**Check type:** `yaml`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `path` | string | one of path/file/files/pattern | Directory to search in |
| `file` | string | one of path/file/files/pattern | Single file name |
| `files` | list | one of path/file/files/pattern | Explicit list of file names |
| `pattern` | string | one of path/file/files/pattern | Regex pattern to match file names |
| `config-name` | string | no | Drupal config name (used with `path` to locate `<config-name>.yml`) |
| `exclude-pattern` | string | no | Regex pattern — matching files are skipped |
| `values` | list of [KeyValue](#keyvalue-fields) | yes | Assertions to evaluate |

## KeyValue fields

Each entry in `values` supports:

| Field | Type | Description |
|---|---|---|
| `key` | string | Dot-separated path to the YAML key (e.g. `check.interval_days`) |
| `value` | string | Expected value (exact string match) |
| `truthy` | bool | If `true`, assert the value is truthy rather than matching `value` |
| `is-list` | bool | If `true`, treat the value at `key` as a list |
| `optional` | bool | If `true`, skip the assertion when the key is absent |
| `disallowed` | list | Values that must not appear at `key` |
| `allowed` | list | Values that must appear at `key` |

## Example

```yaml
checks:
  yaml:
    - name: File config check
      config-name: update.settings
      path: config/default
      values:
        - key: check.interval_days
          value: "7"
```

Source: `pkg/config/testdata/shipshape.yml`

Multiple assertions in one check:

```yaml
checks:
  yaml:
    - name: Cron and update settings
      path: config/default
      config-name: automated_cron.settings
      values:
        - key: interval
          value: "10800"
        - key: status
          truthy: true
```

## Behaviour

Shipshape locates the target file(s) using the supplied path/file/files/pattern
fields. For each file, it parses the YAML and evaluates every `values` entry.
Any assertion that fails is reported as a breach.

When `config-name` is set alongside `path`, Shipshape looks for
`<path>/<config-name>.yml`.

## Remediation

This check does not support automatic remediation. Update the YAML files
manually.