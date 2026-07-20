# drupal-user-role

Checks that all users assigned a given Drupal role are on an explicit
allow-list. Use this to ensure that privileged roles (for example,
`administrator`) are held only by known accounts.

**Check type:** `drupal-user-role`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `drush-path` | string | no | Path to the Drush binary (default: `vendor/drush/drush/drush`) |
| `alias` | string | no | Drush site alias (e.g. `@prod`) |
| `roles` | list | yes | Role machine names to check |
| `allowed-users` | list of int | no | UIDs that are permitted to hold the listed roles |

## Example

```yaml
checks:
  drupal-user-role:
    - name: Administrator role allow-list
      severity: critical
      roles:
        - administrator
      allowed-users:
        - 1
        - 42
```

## Behaviour

Shipshape queries all users holding any of the listed `roles` via Drush. Any
user whose UID is not in `allowed-users` is reported as a breach.

## Remediation

This check does not support automatic remediation. Remove the role from
unexpected users using `drush user:role:remove` or the Drupal UI.