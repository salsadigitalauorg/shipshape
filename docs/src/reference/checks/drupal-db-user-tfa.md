# drupal-db-user-tfa

Checks that all Drupal user accounts have two-factor authentication (TFA)
configured. Use this to enforce MFA compliance across all accounts in a running
environment.

**Check type:** `drupal-db-user-tfa`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `drush-path` | string | no | Path to the Drush binary (default: `vendor/drush/drush/drush`) |
| `alias` | string | no | Drush site alias (e.g. `@prod`) |

## Example

```yaml
checks:
  drupal-db-user-tfa:
    - name: All users have TFA configured
      severity: critical
```

## Behaviour

Shipshape queries the Drupal database via Drush to retrieve all active user
accounts and checks whether each has a TFA method configured. Any user without
TFA is reported as a breach, including their username and UID.

## Remediation

This check does not support automatic remediation. Direct affected users to
configure TFA via their account settings, or enforce TFA at the role level
using the TFA module's configuration.