# docker:exec

The `docker:exec` connection plugin runs a command inside a running Docker
container via the `docker` CLI (`docker exec <container> <command>`). It
backs [`docker:command`](../collect/docker-command.md), letting audits
introspect a live application container — e.g. running `drush` against a
Drupal site — rather than only the host filesystem.

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| container | The name or ID of the running container to exec into. | Yes | "" |
| command | The command and arguments to run. Normally left unset here and supplied instead by the consuming collect plugin's own `command` field (see [`docker:command`](../collect/docker-command.md)). | No | [] |

## Example

```yaml
connections:
  docker-cli:
    docker:exec:
      container: test-shipshape

collect:
  tfa-config:
    docker:command:
      connection: docker-cli
      command: ["/app/vendor/bin/drush", "config:get", "tfa.settings"]
```

Source: `examples/drush-over-docker.yml`

## Requires the `docker` CLI on PATH

`Run` shells out to `docker exec` via `command.ShellCommander` — it does
not talk to the Docker daemon's API directly. The `docker` binary must be
on the shipshape process's `PATH`, and that process must have permission to
use it (typically membership of the `docker` group, or running as root).
Where shipshape itself runs inside a container without Docker CLI access,
this connection cannot be used.

## Errors

| Condition | Behaviour |
| --- | --- |
| `docker` binary not found | `exec.Command` returns a `PathError`; surfaced by the consuming collect plugin (`docker:command`) as a collection error. |
| Container not found, not running, or the command exits non-zero | `exec.ExitError` with stderr attached; surfaced by `docker:command` combining the extracted message with the raw output. |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage. A
container that isn't running therefore stops the run rather than producing
a breach.
:::