# drupal-db-permissions

Checks role permissions directly from the Drupal database. Use this to assert
that sensitive permissions have not been granted to any role in a running
environment.

**Check type:** `drupal-db-permissions`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `drush-path` | string | no | Path to the Drush binary (default: `vendor/drush/drush/drush`) |
| `alias` | string | no | Drush site alias (e.g. `@prod`) |
| `disallowed` | list | no | Permissions that must not be granted to any role |

## Example

```yaml
checks:
  drupal-db-permissions:
    - name: Disallowed permissions
      severity: high
      disallowed:
        - administer config permissions
        - administer modules
        - administer permissions
        - administer seckit
        - administer site configuration
        - administer software updates
        - import configuration
        - synchronize configuration
        - use PHP for google analytics tracking visibility
```

Source: `tests/e2e/suites/shipshape/drupal-permissions.yml`

## Behaviour

Shipshape queries the `role__permissions` table via Drush and checks each role
against the `disallowed` list. Any role that holds a disallowed permission is
reported as a breach, including the role name and the specific permission.

## Remediation

Shipshape can automatically revoke disallowed permissions using
`drush role:perm:remove`. Run Shipshape with the `--remediate` flag to enable
this.