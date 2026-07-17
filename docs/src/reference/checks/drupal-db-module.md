# drupal-db-module

Checks which Drupal modules are enabled or disabled by querying the database
via `drush pm:list`. Use this to verify module state in a running environment
where exported config may not reflect the live database.

**Check type:** `drupal-db-module`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `drush-path` | string | no | Path to the Drush binary (default: `vendor/drush/drush/drush`) |
| `alias` | string | no | Drush site alias (e.g. `@prod`) |
| `required` | list | no | Module names that must be enabled |
| `disallowed` | list | no | Module names that must not be enabled |

## Example

```yaml
checks:
  drupal-db-module:
    - name: Required security modules (live DB)
      required:
        - seckit
        - shield
      disallowed:
        - devel
```

## Behaviour

Shipshape runs `drush [alias] pm:list --status=enabled --format=json` and
parses the output. Any module in `required` that is not enabled is reported as
a breach. Any module in `disallowed` that is enabled is reported as a breach.

## Remediation

This check does not support automatic remediation. Enable or disable modules
using Drush (`drush pm:enable`, `drush pm:uninstall`) or the Drupal UI.