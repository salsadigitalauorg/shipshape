# json:key

The `json:key` collect plugin evaluates an
[RFC 9535](https://www.rfc-editor.org/rfc/rfc9535.html) JSONPath expression
against JSON input and emits the matched values. It is the JSON counterpart to
[`yaml:key`](yaml-key.md), using the JSONPath query language rather than a
dotted-path lookup.

It takes raw JSON as input — typically from [`file:read`](file-read.md) or
[`http:fetch`](http-fetch.md).

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| expression | The RFC 9535 JSONPath expression to evaluate against the input. | Yes | "" |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Return format

The emitted format is derived from the **shape of the match**, mirroring
`yaml:key`:

| Match | Emitted format |
| --- | --- |
| A single scalar | `string` |
| Multiple scalars | `list-string` |
| An object | `map-string` |
| Multiple objects | `list-map-string` |
| No match | `nil` |

Because a non-match emits nil rather than erroring, `not:empty` can be used to
assert the *absence* of a value.

This matters when choosing an analyser: `$.name` on a typical `package.json`
emits a string, so the scalar analysers (`equals`, `not:equals`) apply, whereas
`$.scripts.*` emits a list and needs `allowed:list` or `not:empty`.

## Expression syntax

`json:key` implements RFC 9535 in full, including:

- Wildcards — `$.scripts.*`
- Index and slice — `$.items[0]`, `$.items[1:3]`
- Filters — `$.deps[?@.name=='x'].version` (the RFC form) and the
  parenthesised `[?(@.name=='x')]`
- Functions — `length()`, `count()`, `match()`, `search()`

::: tip Dialect differs from `yaml:key`
`yaml:key` uses an older pre-RFC dialect. It accepts only the parenthesised
filter form and supports none of the RFC functions. An expression written for
`json:key` will not necessarily work on `yaml:key` and vice versa.
:::

## Example

```yaml
collect:
  pkg-file:
    file:read:
      path: package.json

  script-commands:
    json:key:
      input: pkg-file
      expression: "$.scripts.*"

analyse:
  approved-script-tooling:
    allowed:list:
      description: A script uses an unapproved build tool
      input: script-commands
      allowed:
        - eslint .
        - vite build
        - vitest run
```

Source: `examples/json-lookup.yml` — see also the
[`json` recipe](../../guide/gaps.md#recipe-json).

## Errors

| Condition | Behaviour |
| --- | --- |
| `expression` not set | Collection error — `json:key requires an 'expression'` |
| Expression is not valid RFC 9535 | Collection error — `invalid JSONPath expression` |
| Input cannot be decoded as JSON | Collection error — `invalid JSON input` |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage. Malformed
JSON therefore stops the run rather than producing a breach.
:::