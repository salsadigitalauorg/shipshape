# drupal-file-module

Checks which Drupal modules are enabled or disabled by reading the
`core.extension.yml` file from the filesystem. Use this when you want to verify
module state from exported config without a running Drupal instance.

**Check type:** `drupal-file-module`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `path` | string | yes | Directory containing `core.extension.yml` |
| `required` | list | no | Module names that must be enabled |
| `disallowed` | list | no | Module names that must not be enabled |

## Example

```yaml
checks:
  drupal-file-module:
    - name: Required security modules
      path: config/default
      required:
        - seckit
        - shield
      disallowed:
        - devel
        - webprofiler
```

Source: `pkg/config/testdata/shipshape.yml`

## Behaviour

Shipshape reads `<path>/core.extension.yml` and checks the `module` key. Any
module in `required` that is not listed is reported as a breach. Any module in
`disallowed` that is listed is reported as a breach.

## Remediation

This check does not support automatic remediation. Enable or disable modules
using Drush or the Drupal UI, then re-export configuration.