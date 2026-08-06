# http:crawl

The `http:crawl` collect plugin walks a site's link graph starting from a
root url, following same-host links (plus any hosts in `include-domains`)
up to a bounded depth and request count, and returns a map of every
non-2xx response or transport error encountered along the way.

It is a leaf collector (no `input`) - analogous to `http:fetch`, but for
auditing a site's link graph rather than fetching a single resource, e.g.
detecting broken links across a site.

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| url | The root url to start crawling from. | Yes | "" |
| include-domains | Extra hosts to follow links to, in addition to the root url's own host. | No | [] |
| include-urls | Extra paths/urls to fetch regardless of whether they were discovered by crawling - useful for pages not reachable by any link. Resolved relative to `url`. | No | [] |
| max-depth | How many link-following hops from the root url to traverse. The root url itself is depth 0. | No | 3 |
| limit | Hard cap on the total number of requests issued across the whole crawl. | No | 100 |
| timeout | Per-request timeout, as a Go duration string (e.g. `30s`). | No | 30s |
| allow-insecure | Permits `http://` urls. Leave unset for production audits - see the ISM-1139 note below. | No | false |
| resolve-env | Resolve `${VAR}` references in `url` from a `.env` file. See the caveat below. | No | false |
| env-file | Path to the `.env` file used when `resolve-env` is true. | No | `<project-dir>/.env` |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## HTTPS required by default

`https://` is required unless `allow-insecure: true` is set. ISM-1139
recommends encrypting data in transit, and defaulting to plain HTTP for an
operator-supplied url risks silently auditing unencrypted traffic. Set
`allow-insecure` only for local/CI use against a plain-http test server -
never for a production audit target.

## `resolve-env` reads a `.env` file, not the OS environment

`resolve-env` (`pkg/env/envresolver.go`) reads **only** a `.env` file in
the project directory (or the path given by `env-file`) via
`godotenv.Read`. It never consults the shipshape process's own OS
environment (`os.Environ()`). A variable exported in the shell that runs
`shipshape run` will **not** be substituted into `url` - it must be
written to a `.env` file instead. This is the single most likely source of
confusion when parameterising the crawl root across environments.

## Domain allowlist

Only the root url's host, plus any hosts listed in `include-domains`, are
followed - matched as an **exact host string**: subdomains and `www` vs
apex variants of the same site are not implied, so a redirect from
`www.example.gov.au` to `example.gov.au` needs the target added to
`include-domains` explicitly if you want it crawled.

- A **link** (`<a href>`) to a host outside the allowlist is skipped
  entirely: not visited, not recorded, not even as a breach.
- A **redirect** to a host outside the allowlist is refused before it is
  ever issued, enforced on *every* hop of a redirect chain, not only the
  initial request - so the allowlist cannot be bypassed by an on-allowlist
  page redirecting elsewhere. A refusal on a discovered link is logged at
  warning level and skipped - not recorded as a breach - because sites
  commonly redirect off-host for legitimate reasons (SSO, `www`/apex
  normalisation, link-tracking wrappers), and treating every such redirect
  as a broken link would make this check impractical on real sites.
- The one exception is the root `url` itself: if it redirects to a host
  outside the allowlist, the crawl fails outright rather than silently
  visiting nothing and reporting no broken links. Fix by pointing `url` at
  the post-redirect host, or adding the target to `include-domains`.

This keeps a compromised or malicious link (or redirect) on the audited
site from causing shipshape to fetch arbitrary third-party infrastructure
during an audit run.

## Return format

A map of url to status/error, containing **only** non-2xx responses and
transport failures:

| Key | Value |
| --- | --- |
| \<crawled url\> | The HTTP status code as a string (e.g. `"404"`), or `"error: <message>"` for a transport-level failure (DNS, connection refused, timeout). |

2xx responses - including 203 and 204 - are never recorded. A site with no
broken links emits an empty map. Pair the output with the `not:empty`
analyser, which breaches whenever its input is non-empty.

## Example

<<< @/../examples/crawl.yml
