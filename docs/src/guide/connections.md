# Connections

Connections can be used to connect to external systems when gathering data.

## Database

After a connection is defined in the `connections` section of the config file,
it can be used in facts in the `collect` section for plugins that support them.

```yaml{2,12}
connections:
  drupal-db:
    mysql:
      host: mariadb
      user: drupal
      password: drupal
      database: drupal

collect:
  domain-in-tables:
    database:search:
      connection: drupal-db
      search: "%.example.com%"
      id-field: entity_id
```

See the full example [here](https://github.com/salsadigitalauorg/shipshape/blob/1.x/examples/domain-in-db-tables.yml).


## Docker container

```yaml{2,10}
connections:
  docker-cli:
    docker:exec:
      container: test-shipshape

collect:
  tfa-config:
    docker:command:
      connection: docker-cli
      command: ["/app/vendor/bin/drush", "config:get", "tfa.settings"]
  tfa-status:
    yaml:key:
      input: tfa-config
      path: enabled
```
