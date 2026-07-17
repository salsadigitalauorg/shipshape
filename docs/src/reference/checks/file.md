# file

Asserts that no files matching a pattern exist at a given path. Use this to
detect disallowed files — for example, database admin tools committed to the
web root.

**Check type:** `file`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `path` | string | yes | Directory to search in |
| `disallowed-pattern` | string | yes | Regular expression. Any matching filename is a breach |
| `skip-dir` | string | no | Directory name to skip during the search |

## Example

```yaml
checks:
  file:
    - name: Illegal files
      severity: high
      path: web
      disallowed-pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'
```

Source: `tests/e2e/suites/shipshape/files-illegal.yml`

## Behaviour

Shipshape walks `path` recursively. Any file whose name matches
`disallowed-pattern` is reported as a breach. The check passes if no matching
files are found.

Set `skip-dir` to exclude a subdirectory from the walk — for example, to
ignore a vendor directory that legitimately contains files matching the
pattern.

## Remediation

This check does not support automatic remediation. Remove disallowed files
manually.
