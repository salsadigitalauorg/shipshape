# docker:command

The `docker:command` collect plugin runs a command inside a running Docker
container, via a [`docker:exec`](../connection/docker-exec.md) connection,
and emits its output. It is commonly used to run `drush` or similar
in-container tooling against a live application container rather than the
host filesystem.

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| command | The command and its arguments to run inside the container, as a list. | Yes | [] |
| as-list | Split the command's output into a list of lines rather than returning it as raw bytes. | No | false |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Connections

| Connection plugin | Support level |
| --- | :---: |
| [`docker:exec`](../connection/docker-exec.md) | Required |

## Return format

| `as-list` | Format |
| --- | --- |
| `false` (default) | `raw` — the command's stdout as bytes, suitable as input to `yaml:key`, `json:key`, etc. |
| `true` | `list-string` — stdout split into lines. |

## Example

Reading a Drupal site's TFA (two-factor authentication) configuration via
`drush` inside a running container, then checking it is enabled:

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
  tfa-status:
    yaml:key:
      input: tfa-config
      path: enabled
```

Source: `examples/drush-over-docker.yml`

## Errors

| Condition | Behaviour |
| --- | --- |
| The command exits non-zero, or the container/binary cannot be found | Collection error combining the extracted stderr message (via `command.GetMsgFromCommandError`) with the raw output. |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage. A
missing `drush` binary or a container that isn't running therefore stops the
run rather than producing a breach.
:::