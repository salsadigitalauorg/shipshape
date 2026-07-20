# phpstan

Runs [PHPStan](https://phpstan.org/) static analysis and fails if any errors
are reported. Use this to enforce code quality standards as part of a
Shipshape audit.

**Check type:** `phpstan`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `binary` | string | no | Path to the PHPStan binary (default: auto-detected) |
| `configuration` | string | no | Path to the PHPStan config file (e.g. `phpstan.neon`) |
| `paths` | list | yes | Directories or files to analyse |

## Example

```yaml
checks:
  phpstan:
    - name: PHPStan analysis
      configuration: phpstan.neon
      paths:
        - web/modules/custom
        - web/themes/custom
```

Source: `tests/e2e/suites/shipshape/phpstan.yml`

## Behaviour

Shipshape runs PHPStan with `--error-format=json` and parses the output. Any
file errors or global errors reported by PHPStan are recorded as breaches,
including the file path, line number, and error message.

## Remediation

This check does not support automatic remediation. Fix the reported PHPStan
errors in your PHP source code.