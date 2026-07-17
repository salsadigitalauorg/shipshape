# crawler

Crawls a website and reports any page that does not return HTTP 200. Use this
to detect broken pages or missing routes before a deployment is considered
healthy.

**Check type:** `crawler`

## Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label shown in output |
| `severity` | string | no | `low`, `normal`, `high`, or `critical` (default: `normal`) |
| `domain` | string | yes | Base URL to crawl (e.g. `https://example.com`) |
| `extra_domains` | list | no | Additional domains to follow links to |
| `include_urls` | list | no | Specific URLs to include in the crawl regardless of discovery |
| `limit` | int | no | Maximum number of pages to crawl (0 = unlimited) |

## Example

```yaml
checks:
  crawler:
    - name: Site health check
      domain: https://example.com
      limit: 500
```

## Behaviour

Shipshape uses [Colly](https://github.com/gocolly/colly) to crawl the site
starting from `domain`. It follows links that stay within `domain` and any
`extra_domains`. Any URL that returns a non-200 HTTP status is reported as a
breach, including the URL and the status code received.

## Remediation

This check does not support automatic remediation. Investigate and fix the
pages returning non-200 responses.