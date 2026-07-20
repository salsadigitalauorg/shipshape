# drupal-admin-user

Checks that the Drupal UID 1 account has only permitted roles. UID 1 is the
superuser account and should not hold roles that grant broad administrative
access in production.

**Check type:** `drupal-admin-user`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `drush-path` | string | no | Path to the Drush binary (default: `vendor/drush/drush/drush`) |
| `alias` | string | no | Drush site alias (e.g. `@prod`) |
| `allowed-roles` | list | no | Roles that UID 1 is permitted to hold. Any other role is a breach |

## Example

```yaml
checks:
  drupal-admin-user:
    - name: UID 1 roles
      severity: critical
      allowed-roles:
        - authenticated
```

Source: `tests/e2e/suites/shipshape/drupal-role-isadmin.yml`

## Behaviour

Shipshape runs `drush [alias] user:information 1 --format=json` and inspects
the roles assigned to UID 1. Any role not in `allowed-roles` is reported as a
breach.

## Remediation

Shipshape can automatically remove disallowed roles from UID 1 using
`drush user:role:remove`. Run Shipshape with the `--remediate` flag to enable
this.