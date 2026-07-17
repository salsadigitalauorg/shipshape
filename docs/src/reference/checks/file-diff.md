# file:diff

Compares a file on disk against a reference template and reports any
differences. Use this to ensure configuration files have not drifted from a
known-good baseline.

**Check type:** `file:diff`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `target` | string | yes | Path to the file being checked |
| `source` | string | yes | Path to the reference (template) file |
| `source-context` | map | no | Variables injected into the template at render time |
| `context-lines` | int | no | Number of context lines shown around each diff hunk (default: 3) |
| `ignore-missing` | bool | no | If `true`, pass silently when `target` does not exist |

## Example

```yaml
checks:
  file:diff:
    - name: Nginx config matches template
      target: /etc/nginx/nginx.conf
      source: templates/nginx.conf.j2
      context-lines: 5
```

## Behaviour

The `source` file is rendered as a [Gonja](https://github.com/nikolalohinski/gonja)
template with any values from `source-context` available as template variables.
The rendered output is then compared against `target` using a unified diff. Any
difference is reported as a breach.

Set `ignore-missing: true` when the target file is optional — for example, when
checking an environment-specific override that may not exist in all
environments.

## Remediation

This check does not support automatic remediation. Update the target file
manually to match the template.