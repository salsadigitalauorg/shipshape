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
| `exclude-pattern` | string | no | Regex pattern — matching files are skipped |
| `ignore-missing` | bool | no | If `true`, pass instead of failing when a file or path does not exist |
| `values` | list of [KeyValue](#keyvalue-fields) | yes | Assertions to evaluate |

::: warning `config-name` is not supported by this check
`config-name` belongs to the Drush-based Drupal checks
([`drush-yaml`](drupal-drush-yaml.md) and friends), **not** to `yaml`. Setting it
here has no effect: unknown YAML fields are ignored, so a `yaml` check with only
`config-name` and `path` resolves no file at all and breaches with
`no file provided`. Use `file:` (or `files:` / `pattern:`) to name the target.
:::

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
      path: config/default
      file: update.settings.yml
      values:
        - key: check.interval_days
          value: "7"
```

Multiple assertions in one check:

```yaml
checks:
  yaml:
    - name: Cron and update settings
      path: config/default
      file: automated_cron.settings.yml
      values:
        - key: interval
          value: "10800"
        - key: status
          truthy: true
```

Across many files by pattern:

```yaml
checks:
  yaml:
    - name: No disallowed permissions on any role
      path: config/default
      pattern: '^user\.role\..*\.yml$'
      values:
        - key: permissions
          is-list: true
          disallowed:
            - administer site configuration
```

## Behaviour

Shipshape locates the target file(s) using the supplied path/file/files/pattern
fields, in that precedence order — `file` wins over `files`, which wins over
`pattern`. For each file it parses the YAML and evaluates every `values` entry.
Any assertion that fails is reported as a breach.

If none of `file`, `files` or `pattern` is set, the check breaches with
`no file provided` rather than scanning `path` wholesale.

## Remediation

This check does not support automatic remediation. Update the YAML files
manually.