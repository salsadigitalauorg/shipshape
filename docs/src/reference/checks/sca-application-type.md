# sca:application_type

Detects application frameworks present in a codebase and fails if any
disallowed framework is found above a configurable threshold. Use this to
prevent accidental introduction of a second framework into a project.

**Check type:** `sca:application_type`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `disallowed` | list | yes | Framework names that must not be detected |
| `threshold` | int | no | Minimum number of marker matches before a framework is considered present (default: 1) |
| `entrypoint` | string | no | Root directory to start scanning from |
| `paths` | list | no | Specific paths to scan |
| `markers` | map | no | Custom marker strings per framework name |
| `dirs` | map | no | Directories that indicate a framework's presence |
| `dependencies` | map | no | Dependency names (from composer.json / package.json) that indicate a framework |

## Example

```yaml
checks:
  sca:application_type:
    - name: No Symfony in Drupal project
      disallowed:
        - symfony
      threshold: 20
```

## Behaviour

Shipshape scans the codebase for markers associated with each framework in
`disallowed`. A framework is considered present when the number of marker
matches meets or exceeds `threshold`. Any detected disallowed framework is
reported as a breach.

## Remediation

This check does not support automatic remediation. Remove the disallowed
framework's code and dependencies manually.