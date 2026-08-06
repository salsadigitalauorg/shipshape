# http:fetch

The `http:fetch` collect plugin retrieves the content of a remote URL over HTTPS
and emits it as raw bytes. It is a leaf collector (no `input`) — analogous to
[`file:read`](file-read.md) but for remote sources, e.g. fetching a template file
from a public repository to compare against a provisioned project's copy with
[`file:drift`](file-drift.md).

For crawling a site's link graph rather than fetching a single resource, use
[`http:crawl`](http-crawl.md).

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| url | The absolute url to fetch. Must include a scheme and host. | Yes | "" |
| allow-insecure | Permits `http://` urls. Leave unset for production audits — see the ISM-1139 note below. | No | false |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Return format

`raw` — the response body as bytes. Suitable as input to `file:drift`,
`yaml:key`, or `json:key`.

## HTTPS required by default

`https://` is required unless `allow-insecure: true` is set. ISM-1139 recommends
encrypting data in transit, and defaulting to plain HTTP for an
operator-supplied url risks silently sending audit traffic unencrypted. Set
`allow-insecure` only for local or CI use against a plain-http test server —
never for a production target.

## Only unauthenticated URLs are supported

`http:fetch` sends no credentials. Fetching a file from a **private** repository
is not supported — clone it locally and use [`file:read`](file-read.md) instead.

The error message for a failed fetch says so explicitly, because GitHub returns
`404` (not `401`/`403`) for an unauthorised private repository, which otherwise
looks like a wrong path rather than a permissions problem.

## Example

```yaml
collect:
  template:
    http:fetch:
      url: https://raw.githubusercontent.com/client/project-template/main/.github/workflows/ci.yml

  current:
    file:read:
      path: .github/workflows/ci.yml

  ci-drift:
    file:drift:
      input: template
      additional-inputs: [current]

analyse:
  ci-matches-template:
    drift:
      description: CI workflow has not drifted from the project template
      input: ci-drift
```

Source: `examples/file-drift.yml`

## Errors

| Condition | Behaviour |
| --- | --- |
| `url` not set | Collection error — `http:fetch requires a valid, absolute url` |
| `url` has no scheme or host | Collection error — same as above |
| Scheme is not `https` and `allow-insecure` is unset | Collection error — `http:fetch requires an https url; set allow-insecure to permit http` |
| Non-2xx response or transport failure | Collection error, including the private-repository hint |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage. An
unreachable url therefore stops the run rather than producing a breach.
:::