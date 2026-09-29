# database:search

The `database:search` collect plugin searches for a value across text
columns of a database, via a [`mysql`](../connection/mysql.md) connection,
and emits the matching row IDs per table/column. It is used to find
leftover references to a value (e.g. a legacy domain) scattered across an
application's database — most commonly a Drupal site's config/content
tables.

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| search | The value to search for, as a SQL `LIKE` pattern (e.g. `%.example.com%`). | Yes | "" |
| id-field | The column name to select as the row identifier for any match. | Yes | "" |
| tables | A map of table name → list of column names to search. If omitted, every `char`/`varchar`/`longtext`/`longblob` column across the whole database is searched. | No | {} |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Connections

| Connection plugin | Support level |
| --- | :---: |
| [`mysql`](../connection/mysql.md) | Required |

## Return format

`map-nested-string` — table name → column name → comma-separated list of
matching `id-field` values.

## Omitting `tables` scans the whole schema

When `tables` is not set, `database:search` first queries
`information_schema.columns` for every `char`, `varchar`, `longtext`, and
`longblob` column in the target database, then searches each one. This is
convenient for an unfamiliar schema but can be slow on a large database —
set `tables` explicitly to scope the search once the relevant tables are
known.

## A column missing `id-field` is not a fatal error

If a matched column's table does not have the configured `id-field` column,
MySQL's "Unknown column" error for that specific query is swallowed rather
than added to the fact's errors — the search simply continues with the
remaining table/column pairs. Any other database error is recorded via
`AddErrors`.

## Example

```yaml
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
      # Not specifying any tables might be very time-consuming, depending on
      # the database being scanned. Uncomment below to limit to specific tables
      # and fields.
      # tables:
      #   block_content__body:
      #     - body_value
      #   block_content_revision__body:
      #     - body_value

analyse:
  domain-found-in-tables:
    not:empty:
      description: Domain found in table
      input: domain-in-tables
      breach-format:
        type: key-value
        key-label: Table
        key: ' {{ .Breach.Key }}'
        value-label: '[Column: {{ .Breach.ValueLabel }}]'
        value: 'Entity IDs: {{ .Breach.Value }}'
```

Source: `examples/domain-in-db-tables.yml`

## Errors

| Condition | Behaviour |
| --- | --- |
| `id-field` not set | Collection error — `id-field is required`. |
| The connection cannot be established | Collection error from the underlying `database/sql` open/dial. |
| The schema query (`information_schema.columns`) fails, when `tables` is unset | Collection error. |
| A search query fails for a reason other than a missing `id-field` column | Collection error recorded via `AddErrors`; remaining table/column pairs are still searched. |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage.
:::

::: warning Credentials in config
The `mysql` connection's `password` field is a plain string with no
built-in env var substitution or secrets-manager integration — unlike the
`resolve-env` support on some collect plugins. Avoid committing real
credentials to a policy file; inject the password via a config
templating/merge step in your CI pipeline instead, consistent with
ISM-1466 (protecting authentication credentials from disclosure).
:::