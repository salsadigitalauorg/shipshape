# drupal-user-forbidden

Checks that a specific Drupal user account does not exist. Use this to assert
that known-bad accounts (for example, a default admin account with a predictable
username) have been removed.

**Check type:** `drupal-user-forbidden`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `drush-path` | string | no | Path to the Drush binary (default: `vendor/drush/drush/drush`) |
| `alias` | string | no | Drush site alias (e.g. `@prod`) |
| `uid` | string | yes | UID of the user account to check for |

## Example

```yaml
checks:
  drupal-user-forbidden:
    - name: Default admin account removed
      severity: critical
      uid: "1"
```

## Behaviour

Shipshape runs `drush [alias] user:information <uid> --format=json` and checks
whether the account exists. If it does, the check reports a breach including
the username.

## Remediation

Shipshape can automatically block or cancel the forbidden user account using
Drush. Run Shipshape with the `--remediate` flag to enable this.