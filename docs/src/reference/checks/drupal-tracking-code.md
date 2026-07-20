# drupal-tracking-code

Checks that no third-party tracking code (analytics, tag managers) is embedded
in Drupal configuration. Use this to prevent accidental inclusion of tracking
scripts in production environments.

**Check type:** `drupal-tracking-code`

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
  drupal-tracking-code:
    - name: No tracking code in Drupal config
      severity: high
```

## Behaviour

Shipshape fetches the site URI from `drush status` and then scans Drupal
configuration for known tracking code patterns (Google Analytics, Google Tag
Manager, and similar). Any match is reported as a breach.

## Remediation

This check does not support automatic remediation. Remove tracking code from
Drupal configuration via the relevant module settings page or by editing the
exported config files.