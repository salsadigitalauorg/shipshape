# 0.x to 1.x recipe catalogue

1.x does not provide a dedicated plugin for each 0.x check. Instead, you
reproduce a 0.x check by **composing general-purpose `collect` and `analyse`
plugins**. This page is the recipe catalogue: for every 0.x check it shows the
plugin chain that achieves the same outcome in 1.x.

See [The 1.x way: compose, don't port](versions.md#the-1-x-way-compose-don-t-port)
for the reasoning behind this approach.

::: tip Reading this table
None of the statuses below mean "waiting for a one-to-one port". Almost every
0.x check is reproducible today by composing plugins that already exist.

- **Achievable now** — a documented recipe and a working example in `examples/`
  exist. Copy the example and adapt it.
- **Achievable, undocumented** — the same plugins reproduce the check, but no
  worked example is published yet. The plugin chain is listed so you can build
  it; a published recipe is [on the roadmap](roadmap.md).
- **Needs new capability** — no existing building block collects the required
  data. The 1.x direction is still to be determined (TBD).
:::

## File checks

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`file`](../reference/checks/file.md) | Achievable now | `file:lookup` + `not:empty` — see `examples/files.yml` |
| [`file:diff`](../reference/checks/file-diff.md) | Achievable now | `file:read`/`http:fetch` + `file:drift` + `drift` — see `examples/file-drift.yml` |

### Recipe: file

```yaml
collect:
  disallowed-php-scripts:
    file:lookup:
      path: web
      pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'

analyse:
  disallowed-php-scripts-found:
    not:empty:
      description: Disallowed PHP scripts found
      input: disallowed-php-scripts
```

Source: `examples/files.yml`

### Recipe: file:diff

0.x `file:diff` rendered a Jinja template with an operator-supplied vars map
and diffed the result against an on-disk file. 1.x deliberately does **not**
reproduce that: `file:drift` instead masks placeholder-shaped substrings
(default <code v-pre>{{ VAR }}</code>) out of both sides before diffing, so the
same config runs unmodified across every project provisioned from a template —
no per-project vars map required. It also detects a placeholder that was never
substituted, which rendering cannot.

Template rendering was rejected in favour of drift matching for several
reasons: a per-project vars map does not scale across many provisioned
projects; <code v-pre>${{ VAR }}</code>-style placeholders collide with GitHub
Actions' <code v-pre>${{ secrets.* }}</code> syntax when the template is a
workflow file; rendering silently substitutes undefined variables rather than
surfacing them; rendering cannot detect a placeholder that was never
substituted, because there is nothing left to compare once it is rendered
away; and a rendering engine (e.g. gonja) pulls in dozens of transitive
dependencies where RE2 (already in the standard library) needs none.

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

`drift` breaches carry a full unified diff (3 lines of context) via
`KeyValuesBreach` — legible in `pretty`, `json`, and `junit` output. The
`table` renderer does not wrap multi-line breach values, so it is unsuitable
for drift checks.

**Caveat — placeholder matching is a wildcard, not a value check.** Each
placeholder in a template line is matched against the current line with a
non-greedy `(.*?)` capture, so `file:drift` confirms *something* occupies the
placeholder's position, not that the surrounding line is otherwise identical
in structure. In particular:

- A placeholder capture can absorb adjacent genuine drift on the same line —
  e.g. template <code v-pre>image: {{ IMG }}:v1</code> against current
  `image: evil/malware:latest:v1` reports no drift, because the wildcard
  swallows everything up to the literal `:v1` suffix. Keep placeholders on
  their own line, or as the entire value, where the drift you care about
  could plausibly appear.
- The same placeholder name used twice on one line (e.g.
  <code v-pre>a: {{ X }} b: {{ X }}</code>) is **not** checked for consistency
  between the two captured values — RE2 has no backreferences, so `file:drift`
  cannot express "these two captures must match". A current line of
  `a: one b: two` reports no drift even though the two `X` values differ.
- An empty substitution (the current file has nothing where the template has
  <code v-pre>{{ VAR }}</code>) counts as substituted, not as drift. Only a
  placeholder that is still **literally present** in the current file — the
  common copy-paste-and-forgot-to-fill-in bug — is flagged as unsubstituted.

None of these are considered blocking for the initial recipe: they are
inherent to wildcard-based matching without a rendering/vars step, and are
judged an acceptable trade-off against the config-per-project cost of true
template rendering (see the rationale above). Prefer isolating a placeholder to its
own line, or to the entirety of a value, where precision matters.

Source: `examples/file-drift.yml`

## YAML / JSON checks

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`yaml`](../reference/checks/yaml.md) | Achievable now | `file:read` + `yaml:key` + `equals` / `allowed:list` — see `examples/drupal-config.yml` |
| [`yamllint`](../reference/checks/yaml-lint.md) | Achievable now | `file:lookup` + `yaml:lint` + `not:empty` — see `examples/yaml-lint.yml` |
| [`json`](../reference/checks/json.md) | Achievable now | `file:read` + `json:key` + `equals` / `allowed:list` — see `examples/json-lookup.yml` |

### Recipe: yaml

```yaml
collect:
  extension-file:
    file:read:
      path: core.extension.yml

  modules:
    yaml:key:
      input: extension-file
      path: module
      keys-only: true

analyse:
  required-modules:
    allowed:list:
      description: Required modules are not enabled
      input: modules
      required:
        - seckit
        - lagoon_logs
```

Source: `examples/drupal-config.yml`

### Recipe: yamllint

`yamllint` asserts only that files *parse*, reporting an undecodable file as a
breach. Reproducing it needed a new building block, because the 0.x check and
the 1.x pipeline were structurally opposed: in 1.x a YAML parse failure is a
**collect error**, and any collect error is fatal to the whole run —
`fact.Manager().CollectAllFacts()` is followed immediately by
`log.Fatal("failed to collect facts")` (`pkg/shipshape/shipshape.go:171-174`).
So the exact condition `yamllint` exists to report was the condition that
prevented 1.x from reaching the analyse stage at all.

[`yaml:lint`](../reference/collect/yaml-lint.md) resolves this by inverting the
relationship: it treats a parse failure as ordinary **data** — a
filename-to-error map — rather than as a collection error. `not:empty` then
breaches once per entry.

```yaml
collect:
  config-files:
    file:lookup:
      path: config/default
      pattern: '.*\.yml$'
      file-names-only: false   # emit contents (map-bytes), not names

  invalid-yaml:
    yaml:lint:
      input: config-files

analyse:
  yaml-is-valid:
    not:empty:
      description: 'YAML files that do not parse'
      input: invalid-yaml
      severity: high
      breach-format:
        type: key-value
        key-label: file
        key: '{{ .Breach.Key }}'
        value-label: 'YAML error'
        value: '{{ .Breach.Value }}'
```

Source: `examples/yaml-lint.yml`

File selection is delegated to `file:lookup`, so `path`, `pattern`,
`exclude-pattern` and `skip-dirs` are inherited rather than reimplemented — but
note `file-names-only: false` is required, since `yaml:lint` needs file contents
rather than names.

An alternative design — an opt-in "tolerate collect errors" mode surfacing the
error via `BaseAnalyser`'s existing `p.input.GetErrors()` handling
(`pkg/analyse/base.go:92`) — was rejected. A tolerated collect error is
indistinguishable from genuine misconfiguration: a missing file, an unresolvable
connection and a malformed document would all arrive as the same "input failure",
so an operator could not tell the assertion they asked for from a broken config.
It would also have changed run semantics for every fact, where `yaml:lint` is
additive.

Three differences from the 0.x check are worth knowing:

- **Multi-document files are fully validated.** 0.x used a single
  `yaml.Unmarshal`, which stops after the first document, so a file whose later
  documents were malformed passed silently. `yaml:lint` validates every
  document by default; set `all-documents: false` for 0.x parity.
- **No per-file passes.** 0.x added `<file> has valid yaml.` for each valid
  file. 1.x analysers only emit breaches, so a clean run is a single pass on the
  analyser rather than one per file.
- **A missing path is still fatal.** `file:lookup` requires a `pattern` and
  errors on a nonexistent path, so there is no `ignore-missing` equivalent.

Duplicate keys are reported (as `cannot decode yaml: …`, preserving the 0.x
label for a `yaml.TypeError`), which is worth knowing because a duplicate key
silently overrides the earlier value at runtime.

### Recipe: json

`json:key` is the JSON counterpart to `yaml:key`. It reads raw JSON (typically
from a `file:read`) and evaluates an [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535.html)
JSONPath expression against it — e.g. `$.name`, `$.items[0]`, the wildcard
`$.scripts.*`, or a filter such as `$.deps[?@.name=='x'].version`.

The emitted format follows the shape of the match, mirroring `yaml:key`: a
single scalar emits a string, multiple values emit a list, and objects emit a
map. An expression that matches nothing emits nil rather than erroring, so
`not-empty` can act on the absence of a value.

Note that `json:key` implements RFC 9535, whereas `yaml:key` uses an older
pre-RFC dialect. The practical difference is filter syntax: `json:key` accepts
both `[?@.x=='y']` (the RFC form) and `[?(@.x=='y')]`, and supports the RFC
functions `length()`, `count()`, `match()` and `search()`; `yaml:key` accepts
only the parenthesised form and none of the functions.

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

Source: `examples/json-lookup.yml`

## Drupal checks

Every Drush-based Drupal check follows the same pattern: run a `command` (or
`docker:command`) to fetch data, parse it with `yaml:key` where needed, then
assert with `allowed:list` or `equals`. See the
[worked recipe](#recipe-drupal-admin-user-the-composition-pattern) below.

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`drush-yaml`](../reference/checks/drupal-drush-yaml.md) | Achievable now | `docker:command` + `yaml:key` + `equals` — see `examples/drush-over-docker.yml` |
| [`drupal-file-module`](../reference/checks/drupal-file-module.md) | Achievable now | `file:read` + `yaml:key` + `allowed:list` — see `examples/drupal-config.yml` |
| [`drupal-db-module`](../reference/checks/drupal-db-module.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-db-module.yml` |
| [`drupal-db-permissions`](../reference/checks/drupal-db-permissions.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-db-permissions.yml` |
| [`drupal-db-user-tfa`](../reference/checks/drupal-db-user-tfa.md) | Achievable now | `command` + `equals` — see `examples/drupal-db-user-tfa.yml` |
| [`drupal-admin-user`](../reference/checks/drupal-admin-user.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-admin-user.yml` |
| [`drupal-user-forbidden`](../reference/checks/drupal-user-forbidden.md) | Achievable now | `command` + `equals` — see `examples/drupal-user-forbidden.yml` |
| [`drupal-role-permissions`](../reference/checks/drupal-role-permissions.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-role-permissions.yml` |
| [`drupal-user-role`](../reference/checks/drupal-user-role.md) | Achievable now | `command` + `allowed:list` — see `examples/drupal-user-role.yml` |
| [`drupal-tracking-code`](../reference/checks/drupal-tracking-code.md) | Achievable now | `command` + `equals` — see `examples/drupal-tracking-code.yml` |

### Recipe: drush-yaml via Docker

```yaml
connections:
  docker-cli:
    docker:exec:
      container: my-app

collect:
  tfa-config:
    docker:command:
      connection: docker-cli
      command: ["/app/vendor/bin/drush", "config:get", "tfa.settings"]

  tfa-status:
    yaml:key:
      input: tfa-config
      path: enabled

analyse:
  tfa-enabled:
    equals:
      description: TFA must be enabled
      input: tfa-status
      value: "true"
```

Source: `examples/drush-over-docker.yml`

### Recipe: the Drush composition pattern

Every Drush-based check above follows the same shape: run a `command` fact that
emits the data one item per line, then assert with `allowed:list` or `equals`.

```yaml
collect:
  # 1. Collect: run drush and emit the machine name of each super-admin role.
  admin-roles:
    command:
      cmd: bash
      args:
      - -c
      - |
        set -o pipefail
        drush role:list --format=json \
          | jq -r 'to_entries[] | select(.value.is_admin == true) | .key'

analyse:
  # 2. Assert: fail if any role outside the allowed set is present.
  admin-roles-check:
    allowed:list:
      description: Unexpected role carries the is_admin flag
      input: admin-roles
      # `command` output is a map; `key: stdout` selects stdout, which
      # `allowed:list` then splits into one entry per line.
      key: stdout
      allowed:
        - administrator
```

Source: `examples/drupal-admin-user.yml`

Swap the `command` and the `allowed:list` / `equals` assertion to reproduce any
of the other Drush-based checks — the shape stays the same.

## Code quality checks

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`phpstan`](../reference/checks/phpstan.md) | Achievable now | `static-analysis` + `static-analysis:breaches` — see `examples/phpstan.yml` |
| [`sca:application_type`](../reference/checks/sca-application-type.md) | Achievable now | `file:fingerprint` + `detected` — see `examples/app-type.yml` |

### Recipe: phpstan

```yaml
collect:
  phpstan-check:
    static-analysis:
      tool: phpstan
      config: phpstan.neon
      paths: [web/modules/custom]

analyse:
  phpstan-issues:
    static-analysis:breaches:
      description: PHPStan found code quality issues
      input: phpstan-check
      max-issues: 0
```

Source: `examples/phpstan.yml`

### Recipe: sca:application_type

0.x `sca:application_type` scored each disallowed framework with a weighted
likelihood — markers (content matches in entrypoint files), dirs (directory
presence), and dependencies (manifest lookup) each added to a running total,
which breached once it exceeded a threshold. `file:fingerprint` reproduces
this weighted model rather than simple presence matching (`detect` implies
boolean presence; `fingerprint` conveys accumulated weighted evidence), and
`detected` breaches once per label present in its map-shaped input.

No framework knowledge is baked into the binary: every signature (which
markers/dirs/dependencies identify which label) is entirely operator config
under `frameworks`. The binary only supplies the generic matching mechanism.

```yaml
collect:
  app-type:
    file:fingerprint:
      path: .
      threshold: 20
      entrypoints: ["index.php"]
      weights: {markers: 5, dirs: 5, dependencies: 10}
      manifest: composer.json
      dependency-paths:            # optional override; sensible defaults ship
        - "$.require"
        - "$['require-dev']"
      frameworks:
        drupal:
          markers: ["use Drupal\\Core\\DrupalKernel;"]
          dirs: [web, docroot]
          dependencies: [drupal/core-recommended]
        wordpress:
          markers: [" * @package WordPress"]
          dirs: [wp-content]

analyse:
  no-disallowed-frameworks:
    detected:
      description: Disallowed application framework detected
      input: app-type
      severity: high
```

Source: `examples/app-type.yml`

**Caveat — markers/dirs accumulate per match, dependencies does not.** A
framework with 2 markers each found in 2 entrypoint files scores
`weights.markers * 2 * 2`, not a flat `weights.markers` — and a framework
with the same directory name matched twice (or two configured dir names both
present) scores `weights.dirs` twice. This reproduces 0.x's
`sca:application_type` accumulation model exactly
(`pkg/checks/sca/apptypecheck.go`), so a ported 0.x config scores
identically. Dependencies is the one signal that does **not** accumulate:
however many configured dependency names match, or however many
`dependency-paths` expressions surface them, a framework's dependencies
signal contributes `weights.dependencies` at most once — mirroring 0.x's
single bool-driven award for that signal. Do not assume all three signals
behave the same way when tuning weights and thresholds.

**Caveat — threshold is a "strictly greater than" boundary, and 0 cannot mean
"any single match".** A label is only emitted when its accumulated score is
strictly greater than `threshold`; a score equal to `threshold` does not
breach. Because YAML unmarshals an absent field to Go's zero value,
`threshold: 0` and simply omitting `threshold` are indistinguishable from "use
the default (30)". An operator who genuinely wants "any single match
breaches" cannot express `threshold: 0` and must configure a negative
threshold (e.g. `-1`) instead. This is inherited from 0.x, which has the
identical limitation, and is kept unchanged for compatibility with ported 0.x
threshold values.

**Composing a shared signature library.** Because `frameworks` is plain
operator config, a signature library (a YAML file defining `frameworks` for
common PHP/Node frameworks) can be maintained centrally and composed with a
project-local config via repeated `-f` flags, which support URLs:

```sh
shipshape run . \
  -f https://raw.githubusercontent.com/your-org/shipshape-signatures/main/frameworks.yml \
  -f ./my-audit.yml
```

v2 config deep-merges every `-f` file in order
(`pkg/config/config.go`, `pkg/config/merge.go`). **Caveat: `deepMerge`
replaces slices and scalars, it does not union or append them.** If both the
shared library and the local override define `frameworks.drupal.dirs`, the
local file's list *replaces* the library's list entirely — a local override
adding one directory to an otherwise-good signature must repeat every
existing entry, not just the new one. Every override is logged at `Warn`
with the dotted key path, so a later file silently changing a signature is
never invisible — but it is easy to author an override that unintentionally
drops entries from a signature the author only meant to extend. Only maps
merge recursively; everything else (scalars and slices) is a full
replacement.

**Scope boundary — this recipe answers "is a disallowed framework present?",
not "is this the expected app type?".** `file:fingerprint` emits a map of
label to score for every framework whose score clears the threshold; it does
not emit a fixed list of allowed/expected labels to compare against, so it
cannot directly express "this project must be exactly a Drupal site and
nothing else". Composing the output with `allowed:list` does not work
either: `allowed:list` on a `FormatMapString` input compares the map's
*values* (the scores) against an allow-list, not its keys (the labels) — the
comparison you would actually want. Expected-app-type checking would need a
`labels-only` toggle (emitting the label set rather than label→score) that
`file:fingerprint` does not currently offer. Until then, use this recipe only
for "flag disallowed frameworks", not "assert the one expected framework".

**Sovereignty note.** `file:fingerprint` itself performs local filesystem I/O
only — no network calls. The `-f <url>` signature-library composition
pattern above is outbound network I/O at config-load time, not inside the
fact plugin. For AU-sovereign deployments, host shared signature libraries
in-region (`ap-southeast-2` or `ap-southeast-4`) rather than pulling them from
an offshore URL.

## Infrastructure checks

| 0.x check | Status | 1.x plugin chain |
|---|---|---|
| [`docker:base_image`](../reference/checks/docker-base-image.md) | Achievable now | `docker:images` + `allowed:list` — see `examples/docker.yml` |
| [`crawler`](../reference/checks/crawler.md) | Achievable now | `http:crawl` + `not:empty` — see `examples/crawl.yml` |

### Recipe: docker:base_image

```yaml
collect:
  dockerfile-paths:
    yaml:key:
      input: compose-services-nodes
      path: build.dockerfile

  dockerfiles:
    file:read:multiple:
      input: dockerfile-paths

  base-images:
    docker:images:
      input: dockerfiles
      no-tag: true

analyse:
  disallowed-base-image:
    allowed:list:
      description: Disallowed base image found in Dockerfiles
      input: base-images
      allowed:
        - uselagoon/php-8.2-fpm
        - uselagoon/nginx-drupal
      deprecated:
        - uselagoon/php-8.1-fpm
```

Source: `examples/docker.yml`

### Recipe: crawler

```yaml
collect:
  site-links:
    http:crawl:
      url: ${CRAWL_BASE_URL}
      resolve-env: true
      max-depth: 2

analyse:
  site-has-no-broken-links:
    not:empty:
      input: site-links
      severity: high
      description: "Site has no broken links"
```

Source: `examples/crawl.yml`

**Migration notes from the 0.x `crawler` check:**

| 0.x | 1.x |
|---|---|
| `domain` | `url` |
| `extra_domains` | `include-domains` |
| `include_urls` | `include-urls` |
| `limit: 0` = unlimited | `limit` defaults to `100` — an unbounded crawl is no longer the default; opt into a wider budget explicitly |
| no depth control | `max-depth` (default `3`) |
| 203/204 misreported as errors (a bug in the underlying `colly` v1 library) | only non-2xx responses and transport errors are reported; 203/204 are correctly treated as success |
| robots.txt always ignored | same — this is an audit tool crawling operator-owned/operator-approved targets, not a general-purpose spider |

**Sovereignty note.** `http:crawl` performs outbound network I/O against
the configured `url` (and any `include-domains` hosts) — this is the
intended purpose of the plugin, not an incidental side effect like
`file:fingerprint`'s signature-library composition. Scope is bounded to
the root host plus explicitly allow-listed domains; no credentials are
forwarded with requests; nothing is persisted beyond what an `output:`
sink writes. For AU-sovereign audits, crawl targets should be
operator-controlled or operator-approved infrastructure.

**SSRF note.** `url` (and `include-domains`) must come from an explicitly
configured, trusted value — e.g. a value set in a CI pipeline's own
config, never a value influenced by end-user input. The domain allowlist
bounds lateral scope: it is enforced on both discovered links and
redirects — including every hop of a redirect chain, not only the initial
request — so an attacker who fully controls the audited site's markup
cannot make `http:crawl` fetch arbitrary third-party hosts either by
linking to them directly or by redirecting to them. It does not protect
against a malicious or attacker-influenced `url` itself; if the
configured root redirects outside the allowlist, the crawl fails outright
rather than silently auditing nothing (see the reference page's "Domain
allowlist" section for the link-vs-redirect distinction and why a
discovered link redirecting off-host is logged rather than treated as a
broken link).

**Privacy note (APP 11).** Crawled URLs, including any query strings,
appear verbatim in breach output. An `output:` sink therefore inherits the
same data classification as the URL structure of the audited site — treat
accordingly if the site's URLs can embed identifiers or other information
that should not be duplicated into a lower-classification results store.

## Summary

| Status | Count |
|---|---|
| Achievable now | 19 |
| Achievable, undocumented | 0 |
| Needs new capability | 0 |

19 rows, one per registered 0.x check type. Every 0.x check now has a documented
recipe and a working example in `examples/` — the last gap, `yamllint`, closed
with the [`yaml:lint`](../reference/collect/yaml-lint.md) fact plugin.

Remaining work is documentation quality rather than new capability; see the
[roadmap](roadmap.md).
