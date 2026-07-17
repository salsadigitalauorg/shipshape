# drush-yaml

Reads Drupal configuration via `drush config:get` and asserts key-value pairs.
Use this when you need to check live Drupal config rather than exported YAML
files — for example, in a running environment.

**Check type:** `drush-yaml`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `drush-path` | string | no | Path to the Drush binary (default: `vendor/drush/drush/drush`) |
| `alias` | string | no | Drush site alias (e.g. `@prod`) |
| `config-name` | string | yes | Drupal config object name (e.g. `update.settings`) |
| `command` | string | no | Drush sub-command override (default: `config:get`) |
| `remediate-command` | string | no | Drush command to run when remediating a breach |
| `remediate-msg` | string | no | Message to display after remediation |
| `values` | list of [KeyValue](#keyvalue-fields) | yes | Assertions to evaluate |

## KeyValue fields

Each entry in `values` supports:

| Field | Type | Description |
|---|---|---|
| `key` | string | Dot-separated path to the config key (e.g. `check.interval_days`) |
| `value` | string | Expected value (exact string match) |
| `truthy` | bool | If `true`, assert the value is truthy rather than matching `value` |
| `is-list` | bool | If `true`, treat the value at `key` as a list |
| `optional` | bool | If `true`, skip the assertion when the key is absent |
| `disallowed` | list | Values that must not appear at `key` |
| `allowed` | list | Values that must appear at `key` |

## Example

```yaml
checks:
  drush-yaml:
    - name: Ensure correct update settings
      config-name: update.settings
      values:
        - key: check.interval_days
          value: "7"
```

Source: `pkg/config/testdata/shipshape.yml`

## Behaviour

Shipshape runs `drush [alias] config:get <config-name> --format=yaml` and
parses the output. Each `values` entry is then evaluated against the parsed
YAML. Any assertion that fails is reported as a breach.

## Remediation

Set `remediate-command` to a Drush command that corrects the breach. When
Shipshape runs in remediation mode, it executes that command for each failing
assertion. Set `remediate-msg` to provide a human-readable summary of what was
changed.