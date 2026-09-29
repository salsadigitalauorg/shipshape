# docker:images

The `docker:images` collect plugin parses one or more Dockerfiles and emits
the base image (the `FROM` instruction target) used by each. It is used to
audit which base images a project's Dockerfiles pull, e.g. against an
allow-list via [`allowed:list`](../analyse/allowed-list.md).

It takes a `map-bytes` input — typically [`file:read:multiple`](file-read-multiple.md)
or [`file:lookup`](file-lookup.md) — of Dockerfile content keyed by path.

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| no-tag | Strip the tag (and digest) from each emitted image reference, leaving only the image name and trailing `:latest`. | No | false |
| args-from | The name of the `additional-inputs` entry supplying `ARG` default overrides per Dockerfile, used to resolve `FROM ${ARG}` references. | No | "" |
| ignore | Image references to omit from the output. Entries can reference env vars, resolved against `args-from` when set. | No | [] |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Input formats

| Input format | Support level |
| --- | :---: |
| `map-bytes` | Required |

## Additional inputs

When `args-from` is set, `additional-inputs` must include an entry with
that name, carrying `map-nested-string` data (Dockerfile path → arg name →
value) — typically produced by a `yaml:key` lookup of `build.args` from a
Docker Compose file.

## Return format

`map-list-string` — Dockerfile path to the list of base images it declares
(one entry per `FROM` instruction, e.g. multi-stage builds emit multiple
entries per file).

## `FROM` args are resolved per Dockerfile, in declaration order

Each Dockerfile is parsed independently. `ARG` instructions preceding a
`FROM` populate an args map used to resolve `${ARG_NAME}` in that and
subsequent `FROM` instructions within the same file — mirroring Docker's own
build-time arg scoping. An `ARG` with no default and no `args-from` override
resolves to an empty string, producing an image reference like `:latest`
rather than a collection error — this is a lossy but non-fatal case worth
checking for in `allowed:list` config.

## Example

Extracting a Compose file's per-service Dockerfiles and their base images,
resolving build args from the same Compose file:

```yaml
collect:
  compose-file:
    file:read:
      path: docker-compose.yml

  compose-services-nodes:
    yaml:key:
      input: compose-file
      path: services
      nodes-only: true

  dockerfile-paths:
    yaml:key:
      input: compose-services-nodes
      path: build.dockerfile

  buildargs:
    yaml:key:
      input: compose-services-nodes
      path: build.args
      resolve-env: true

  dockerfiles:
    file:read:multiple:
      input: dockerfile-paths

  base-images:
    docker:images:
      input: dockerfiles
      additional-inputs: [buildargs]
      no-tag: true
      args-from: buildargs
      ignore:
        - $CLI_IMAGE

analyse:
  disallowed-base-image:
    allowed:list:
      description: Disallowed base image found in services' Dockerfiles.
      input: base-images
      allowed:
        - uselagoon/php-8.1-cli-drupal
        - uselagoon/nginx-drupal
      exclude-keys:
        - test
```

Source: `examples/docker.yml`

## Errors

| Condition | Behaviour |
| --- | --- |
| Input format is not `map-bytes` | Collection error — `ErrSupportNone` for `input data format`. |
| `args-from` is set but no matching entry exists in `additional-inputs` | Collection error — `ErrSupportRequired` for `additionalInputs`. |
| A Dockerfile's content is not valid Dockerfile syntax | Collection error from the underlying parser. |
| An `ignore` entry's env var reference cannot be resolved | Collection error. |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage.
:::