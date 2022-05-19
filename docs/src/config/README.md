---
sidebar: auto
---

# Config

The basic layout of the config file is as follows:
```yaml
project-dir: /path/to/project # Default is the current working directory
fail-severity: high # Default is high, other possible values are low, normal, critical
checks:
  {check-type}:
    name: {check-name}
    severity: normal # Only report failures, do not fail
    ... # Other check-specific fields.
```

Taking the config in the quick-start as an example:
```yaml
checks:
  file: # Corresponds to {check-type}
    - name: Illegal files # Corresponds to {check-name}
      path: web
      disallowed-pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'
```

## Check types

The following check types are available:
  - [file](#file)
  - [yaml](#yaml)
  - [yamllint](#yamllint)
  - [crawler](#crawler)
  - [drush-yaml](#drush-yaml)
  - [drupal-file-module](#drupal-file-module)
  - [drupal-db-module](#drupal-db-module)
  - [drupal-db-permissions](#drupal-db-permissions)
  - [phpstan](#phpstan)

### Common fields
The fields below are common to all checks.

| Field    | Default | Required | Description               |
| -------- | :-----: | :------: | ------------------------- |
| name     |    -    |   Yes    | The name of the check     |
| severity | normal  |    No    | The severity of the check |

### file
Checks for disallowed files in the specified path using the pattern provided.

| Field              | Default | Required | Description                                         |
| ------------------ | :-----: | :------: | --------------------------------------------------- |
| path               |    -    |   Yes    | Path (directory) to check for the presence of files |
| disallowed-pattern |    -    |   Yes    | Regex pattern defining the disallowed files         |

#### Example
```yaml
file:
  - name: Illegal files
    path: web
    disallowed-pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'
```

### yaml
Checks yaml files for the presence or absence of required/disallowed values.

| Field           | Default | Required | Description                                             |
| --------------- | :-----: | :------: | ------------------------------------------------------- |
| path            |    -    |   Yes    | Path (directory) to check for the presence of files     |
| file            |    -    |    No    | A single file to check                                  |
| files           |    -    |    No    | A list of files to check                                |
| pattern         |    -    |    No    | Regex pattern defining a list of files to check         |
| exclude-pattern |    -    |    No    | Regex pattern to exclude a list of files from the check |
| ignore-missing  |  false  |    No    | Specify whether a missing file is a fail                |
| values          |    -    |   Yes    | The list of keys and values for the check.              |


#### Values
The list of values can either be simple key/value, e.g
```yaml
values:
  - key: foo
    value: bar
```
where the file being checked is:
```yaml
# file-being-checked.yml
a: b
foo: bar
c: d
```
in which case line `foo: bar` would be detected as a breach.

Or it can be a list of disallowed values, e.g
```yaml
values:
  - key: foo
    is-list: true
    disallowed:
      - zoo
      - zoom
      - paf
```
where the file being checked is:

```yaml
# file-being-checked.yml
a: b
foo:
  - bar
  - baz
  - zoo
  - zoom
c: d
```
in which case lines `- zoo` and `- zoom` would be detected as breaches.

#### Example
```yaml
yaml:
  - name: Validate install profile
    file: core.extension.yml
    ignore-missing: true
    path: config/default
    values:
      - key: profile
        value: govcms
  - name: Disallowed permissions
    pattern: user.role.*.yml
    exclude-pattern: user.role.govcms_site_administrator
    ignore-missing: true
    path: config/default
    values:
      - key: is_admin
        value: false
      - key: permissions
        is-list: true
        disallowed:
          - administer modules
          - administer permissions
          - administer site configuration
          - administer software updates
          - Administer the list of modules that can be managed by others
          - import configuration
          - use PHP for google analytics tracking visibility
  - name: Validate TFA config
    file: tfa.settings.yml
    ignore-missing: true
    path: config/default
    values:
      - key: enabled
        value: 1
      - key: required_roles.authenticated
        value: authenticated
```

### yamllint
documentation coming soon...

### crawler
documentation coming soon...

### drush-yaml
documentation coming soon...

### drupal-file-module
documentation coming soon...

### drupal-db-module
documentation coming soon...

### drupal-db-permissions
documentation coming soon...

### phpstan
documentation coming soon...
