# docker:base_image

Checks that Docker base images used in a project are on an explicit allow-list.
Use this to prevent use of untrusted or deprecated base images.

**Check type:** `docker:base_image`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `allowed` | list | yes | Image names (or prefixes) that are permitted |
| `deprecated` | list | no | Image names that are permitted but should be replaced — reported as warnings |
| `exclude` | list | no | File paths to exclude from the scan |
| `pattern` | list | no | Glob patterns to locate Dockerfile files (default: `**/Dockerfile*`) |
| `paths` | list | no | Directories to search for Dockerfiles |

## Example

```yaml
checks:
  docker:base_image:
    - name: Approved base images
      severity: high
      allowed:
        - uselagoon/php-8.2-fpm
        - uselagoon/nginx-drupal
      deprecated:
        - uselagoon/php-8.1-fpm
```

## Behaviour

Shipshape scans for Dockerfile files using the configured `pattern` and
`paths`. For each `FROM` instruction found, it checks whether the image name
matches an entry in `allowed`. Images in `deprecated` are allowed but reported
separately. Any image not in either list is reported as a breach.

Docker Compose files with `build.dockerfile` entries are also scanned.

## Remediation

This check does not support automatic remediation. Update the `FROM`
instructions in your Dockerfiles to use an approved base image.