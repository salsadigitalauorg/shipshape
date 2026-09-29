# file:fingerprint

The `file:fingerprint` collect plugin scores a set of operator-supplied
framework "signatures" against a codebase and emits the labels whose
accumulated score exceeds a threshold. It is the 1.x replacement for 0.x's
`sca:application_type` check
(`pkg/checks/sca/apptypecheck.go`), reproducing the same weighted-evidence
model rather than simple presence matching, so a ported 0.x config
(markers/dirs/weights/threshold) scores identically.

No framework knowledge is baked into the binary — every signature (which
markers, directories, and dependencies identify which label) is entirely
operator config, shareable across projects via YAML anchors and composed
with local overrides via multiple `-f` flags.

## Plugin fields

| Field | Description | Required | Default |
| --- | --- | :---: | :---: |
| path | The directory to scan, relative to the project directory. | Yes | "" |
| frameworks | A map of label → [signature](#framework-signature-fields) to score against the codebase. | Yes | {} |
| threshold | A framework is only emitted when its accumulated score is strictly **greater than** this value. | No | 30 |
| entrypoints | Basenames of entrypoint files to search for markers in (matched exactly, not by substring). | No | `["index.php"]` |
| weights | Score contribution per matched signal — see [Weights fields](#weights-fields). | No | `{markers: 5, dirs: 5, dependencies: 10}` |
| manifest | The manifest filename (relative to `path`) read for the dependencies signal. Only one manifest is supported per fact instance. | No | `composer.json` |
| dependency-paths | RFC 9535 JSONPath expressions evaluated against the parsed manifest; every matched object's keys become candidate dependency names. | No | `["$.require", "$['require-dev']", "$.dependencies", "$.devDependencies"]` |

### Framework signature fields

Each entry under `frameworks` takes:

| Field | Description |
| --- | --- |
| markers | Exact-line strings searched for within each entrypoint file. Scores `weights.markers` per `(marker, entrypoint file)` pair matched — N markers across M entrypoints score `weights.markers * N * M`. |
| dirs | Directory names searched for anywhere under `path`. Scores `weights.dirs` per matching directory occurrence — the same name nested twice scores twice. |
| dependencies | Manifest package names searched for among the values matched by `dependency-paths`. Scores `weights.dependencies` **once only**, no matter how many configured names or expressions match — this is the one signal that does not accumulate per-hit. |

### Weights fields

| Field | Default |
| --- | --- |
| markers | 5 |
| dirs | 5 |
| dependencies | 10 |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Return format

`map-string` — label to accumulated score (as a string), containing only
labels whose score is strictly greater than `threshold`. Sub-threshold
scores are never included, not even as a zero or low value. Pair the
output with the [`detected`](../analyse/detected.md) analyser, which
breaches on any label present in the map.

## Example

```yaml
collect:
  drupal-project:
    file:fingerprint:
      path: app-type/drupal
      threshold: 15
      entrypoints: ["index.php"]
      frameworks: &x-frameworks
        drupal:
          markers: ["use Drupal\\Core\\DrupalKernel;"]
          dirs: [web, docroot]
          dependencies: [drupal/core-recommended]
        wordpress:
          markers: [" * @package WordPress"]
          dirs: [wp-content]
        symfony:
          markers: ["  return new Kernel($context['APP_ENV'], (bool) $context['APP_DEBUG']);"]
          dirs: [bin, config, public]
          dependencies: [symfony/runtime, symfony/symfony, symfony/framework]
        laravel:
          markers: ["use Illuminate\\Contracts\\Http\\Kernel;"]
          dirs: [config, public]
          dependencies: [laravel/framework, laravel/tinker]

analyse:
  drupal-project-frameworks:
    detected:
      description: Disallowed application framework detected
      input: drupal-project
      severity: high
```

Source: `examples/app-type.yml` — see also the
[`sca:application_type` recipe](../../guide/gaps.md#recipe-sca-application-type).

## Threshold default: 30, not 1

0.x's check code defaults `Threshold` to 30
(`pkg/checks/sca/apptypecheck.go`), but 0.x's own reference doc for
`sca:application_type` claims a default of 1. `file:fingerprint` follows the
**code** (30) — what 0.x actually executed at runtime — not the doc value,
which was never true.

## `threshold: 0` cannot mean "any single match breaches"

Because YAML unmarshals an absent field to Go's zero value, `threshold: 0`
and an unset `threshold` are indistinguishable from "use the default" (30).
An operator who wants "any single match breaches" cannot express
`threshold: 0` — they must use a threshold of `-1` or lower. This limitation
is inherited unchanged from 0.x, kept for config compatibility with ported
0.x definitions.

## Entrypoint matching is exact-basename, not substring

`entrypoints` are matched by exact basename, unlike 0.x's substring `Glob`
helper (which does `strings.Contains(name, match)` and would incorrectly
match e.g. `myindex.phpx` against `index.php`). A config relying on 0.x's
looser matching must list the extra basenames explicitly.

## The dependencies signal is only attempted when needed

The manifest is only read, and `dependency-paths` only parsed and
evaluated, when at least one configured framework actually sets
`dependencies`. A config that never uses the dependencies signal does not
fail just because `path` happens not to contain a `composer.json`.

## Data handling: matched text is never emitted

Breach/fact output never includes the matched marker text, file path, or
matched manifest key/version — only the label and its accumulated score.
A carelessly-written marker cannot echo file content (potentially
containing secrets) into an audit report.

## Errors

| Condition | Behaviour |
| --- | --- |
| `frameworks` is empty | Collection error — `file:fingerprint requires at least one entry under 'frameworks'`. |
| `path` is empty | Collection error — `file:fingerprint requires a 'path'`. |
| A `dependency-paths` expression is not valid RFC 9535 JSONPath | Collection error — `file:fingerprint invalid dependency-paths expression`. |
| The manifest is missing (only when a framework needs the dependencies signal) | Collection error — `file:fingerprint manifest not found`. |
| The manifest exists but cannot be read | Collection error — `file:fingerprint manifest unreadable`. |
| The manifest is not valid JSON | Collection error — `file:fingerprint manifest is not valid JSON`. |
| An entrypoint file cannot be read | Collection error — an unreadable entrypoint would silently under-score any framework whose marker it might have matched, so this aborts the run rather than skip the file. |

::: warning A collection error aborts the run
Any fact error is fatal — the pipeline never reaches the analyse stage.
Under-scoring a present framework is the worst failure mode for an audit
tool, so every error above is treated as fatal rather than a silent
zero/low score.
:::