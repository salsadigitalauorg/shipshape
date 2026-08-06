# Overview

Every plugin registered in the 1.x pipeline is listed below. Pages marked
*(stub)* exist but are not yet written up — the plugin works, the documentation
is the gap. Entries with no link have no page yet at all.

## Connection plugins

The following Connection plugins are available:
  - [mysql](../reference/connection/mysql) *(stub)*
  - [docker:exec](../reference/connection/docker-exec) *(stub)*

## Collect plugins

The following Collect/Fact plugins are available:
  - [command](../reference/collect/command)
  - [database:search](../reference/collect/database-search) *(stub)*
  - [docker:command](../reference/collect/docker-command) *(stub)*
  - [docker:images](../reference/collect/docker-images) *(stub)*
  - [file:drift](../reference/collect/file-drift) — see also the [`filediff` recipe](../guide/gaps.md#recipe-file-diff)
  - [file:fingerprint](../reference/collect/file-fingerprint) *(stub)*
  - [file:lookup](../reference/collect/file-lookup) *(stub)*
  - [file:read](../reference/collect/file-read) *(stub)*
  - [file:read:multiple](../reference/collect/file-read-multiple) *(stub)*
  - [http:crawl](../reference/collect/http-crawl)
  - [http:fetch](../reference/collect/http-fetch)
  - [json:key](../reference/collect/json-key) — see also the [`json` recipe](../guide/gaps.md#recipe-json)
  - [static-analysis](../reference/collect/static-analysis)
  - [yaml:key](../reference/collect/yaml-key) *(stub)*

## Analyse plugins

The following Analyse plugins are available:
  - [allowed:list](../reference/analyse/allowed-list)
  - [detected](../reference/analyse/detected) — see also the [`sca:application_type` recipe](../guide/gaps.md#recipe-sca-application-type)
  - [drift](../reference/analyse/drift) — see also the [`filediff` recipe](../guide/gaps.md#recipe-file-diff)
  - [equals](../reference/analyse/equals)
  - [not:empty](../reference/analyse/not-empty)
  - [not:equals](../reference/analyse/not-equals)
  - [regex:match](../reference/analyse/regex-match)
  - [regex:not-match](../reference/analyse/regex-not-match)
  - [static-analysis:breaches](../reference/analyse/static-analysis)

## Remediation plugins

The following Remediation plugins are available:
  - [command](../reference/remediate/command)

## Checks (0.x)

The legacy `checks:` format is documented in the
[Checks reference](../reference/checks/) and the
[0.x config guide](../guide/0.x.md).
