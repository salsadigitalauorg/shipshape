# detected

The `detected` analyser breaches **once per entry** in a map-shaped input,
whenever that entry is present. It is designed to pair with the
[`file:fingerprint`](../collect/file-fingerprint.md) fact, where each map entry is
a detected framework and its likelihood score.

## Plugin fields

`detected` has no plugin-specific fields — there is nothing to configure beyond
the common ones below.

::: tip No inverse polarity
Like [`drift`](drift.md), `detected` has no toggle to invert the assertion. The
presence of an entry is itself the assertion failure, so an empty input passes.
:::

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/analyse.html').key"/>

## Input format

Requires `map-string` (framework label → score). An empty map produces no
breaches. Other formats are treated as a no-op.

## Breach output

One `KeyValueBreach` per map entry, labelled `framework` / `likelihood`.

Map keys are **sorted** before breaches are emitted. Go randomises map iteration
order, and this analyser is typically used with multi-label input (several
detected frameworks), so without sorting the breach order — and any test
asserting on it — would be non-deterministic.

## Example

```yaml
analyse:
  application-type:
    detected:
      description: Detected application frameworks
      input: app-fingerprint
```

Source: `examples/app-type.yml` — see also the
[`sca:application_type` recipe](../../guide/gaps.md#recipe-sca-application-type).