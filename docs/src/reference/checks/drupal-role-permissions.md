# drupal-role-permissions

Checks the permissions assigned to a specific Drupal role. Use this to assert
that a role has all required permissions and none of the disallowed ones.

**Check type:** `drupal-role-permissions`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `drush-path` | string | no | Path to the Drush binary (default: `vendor/drush/drush/drush`) |
| `alias` | string | no | Drush site alias (e.g. `@prod`) |
| `rid` | string | yes | Role machine name to check (e.g. `editor`, `administrator`) |
| `required-permissions` | list | no | Permissions that must be granted to the role |
| `disallowed-permissions` | list | no | Permissions that must not be granted to the role |

## Example

```yaml
checks:
  drupal-role-permissions:
    - name: Editor role permissions
      rid: editor
      required-permissions:
        - access content
        - create article content
      disallowed-permissions:
        - administer users
        - administer permissions
```

## Behaviour

Shipshape queries the role's permissions via Drush. Any permission in
`required-permissions` that is absent is reported as a breach. Any permission
in `disallowed-permissions` that is present is reported as a breach.

## Remediation

This check does not support automatic remediation. Update role permissions via
the Drupal UI or using `drush role:perm:add` / `drush role:perm:remove`.