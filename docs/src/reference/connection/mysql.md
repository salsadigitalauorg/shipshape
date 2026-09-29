# mysql

The `mysql` connection plugin opens a MySQL/MariaDB database connection for
use by collect plugins that need direct database access, primarily
[`database:search`](../collect/database-search.md).

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| host | The database host. | Yes | "" |
| port | The database port. | No | "3306" |
| user | The database user. | Yes | "" |
| password | The database password. | Yes | "" |
| database | The database name. | Yes | "" |

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
```

Source: `examples/domain-in-db-tables.yml`

## Driver and dialect

Connects via [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
and wraps the connection with [goqu](https://github.com/doug-martin/goqu)'s
`mysql` dialect, giving consuming collect plugins (currently
`database:search`) a query builder rather than raw SQL strings.

## No TLS configuration is exposed

The DSN built by this plugin does not set `tls=` — the connection is
plaintext by default, matching `go-sql-driver/mysql`'s own default. There
is currently no plugin field to require TLS to the database server. Where
the database is not reachable only over a private/trusted network, terminate
TLS at the network layer (e.g. a Lagoon/Kubernetes service mesh, an SSH
tunnel, or a cloud provider's managed-TLS proxy) rather than relying on this
connection for transport security, consistent with ISM-0459 (protecting
data in transit).

## Credentials are plain YAML, with no substitution support

`password` (and `user`) are read as literal strings from the parsed YAML —
there is no `${VAR}` substitution or secrets-manager integration built into
this connection. Avoid committing real credentials to a policy file
committed to version control; inject them via a templating/merge step in
CI instead (see the [config guide](../../guide/) for the general `-f`
multi-file merge mechanism), consistent with ISM-1466 (protecting
authentication credentials from disclosure).

## Errors

| Condition | Behaviour |
| --- | --- |
| The `database/sql` driver fails to open the connection (malformed DSN) | Error returned from `Run`, surfaced by the consuming collect plugin as a collection error. |

Connectivity itself (host unreachable, bad credentials, unknown database)
is not verified by `Run` — `sql.Open` only validates the DSN; the actual
connection attempt happens on the first query issued by the consuming
collect plugin.