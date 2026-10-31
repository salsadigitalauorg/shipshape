# Quick-start

## Installation

### MacOS

The preferred method is installation via [Homebrew](https://brew.sh/).
```sh
brew install salsadigitalauorg/shipshape/shipshape
```

To pin a specific release, install the versioned formula. The tap keeps the
five most recent versions:
```sh
brew install salsadigitalauorg/shipshape/shipshape@1.0.0-alpha.1.6
```

Versioned formulae are keg-only, so they are not symlinked into your prefix
while the main formula is installed. To switch between them:
```sh
brew unlink shipshape && brew link --overwrite shipshape@1.0.0-alpha.1.6
```

To compile from source instead of using the published binary, build the `1.x`
branch tip. This requires Go and takes a few minutes:
```sh
brew install --HEAD salsadigitalauorg/shipshape/shipshape
```

### Linux

```sh
curl -L -o shipshape https://github.com/salsadigitalauorg/shipshape/releases/latest/download/shipshape-$(uname -s)-$(uname -m)
chmod +x shipshape
mv shipshape /usr/local/bin/shipshape
```

### Docker

Run directly from a docker image:
```sh
docker run --rm ghcr.io/salsadigitalauorg/shipshape:latest shipshape version
```

Or add to your docker image:
```Dockerfile
COPY --from=ghcr.io/salsadigitalauorg/shipshape:latest /usr/local/bin/shipshape /usr/local/bin/shipshape
```

::: tip Which config format?
Shipshape supports two configuration formats. The **1.x pipeline** (`collect` /
`analyse` / `output`) is the recommended format and is used in production. The
**0.x `checks:` format** is fully supported and continues to run without changes.

See [Config versions](versions.md) for a full comparison and a guide on which
to use.
:::

## Usage

The basic layout of the config file is as follows:
```yaml
# Optional - Set up any connection(s) required for
# collecting data (docker, mysql, etc...).
connections:
  ...

# Required - Set up how to collect the data.
collect:
  ...

# Optional - Set up how to analyse the data.
# Leave empty if no analysis is required.
# Run the command `shipshape collect .` in that case.
analyse:
  ...

# Optional - Configure output format and other output plugin options.
# Defaults to stdout with the pretty format.
output:
  stdout:
    format: pretty  # Format for stdout output
  file:
    path: results/output.xml  # File to write output to
    format: junit  # Format for file output
```

Create a config file:
```yaml
# shipshape.yml
collect:
  disallowed-php-scripts:
    file:lookup:
      path: web
      pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'
  sensitive-public-files:
    file:lookup:
      path: web/sites/default/files
      pattern: '.*\.(sql|php|sh|py|bz2|gz|tar|tgz|zip)?$'
      exclude-pattern: '.*\.(css|js)\.gz?$'
      skip-dir:
        - private

analyse:
  disallowed-php-scripts-found:
    not:empty:
      description: 'Disallowed php scripts found'
      input: disallowed-php-scripts
      severity: high
  sensitive-public-files-found:
    not:empty:
      description: 'Sensitive files found in public directory'
      input: sensitive-public-files
      severity: high
      message: 'Please check the directory web/sites/default/files and remove all sensitive files'
```

Execute the policy:
```sh
shipshape run .
```

```sh
$ shipshape run -h
Execute policies against the specified directory

Usage:
  shipshape run [dir|.] [flags]

Flags:
  -e, --error-code                               Exit with error code if a failure is
                                                 detected (env: SHIPSHAPE_ERROR_ON_FAILURE)
      --fail-severity string                     The severity level at which the program
                                                 should exit with an error (default "high")
  -h, --help                                     help for run
      --lagoon-clear-problems                    Delete existing Lagoon problems for this run's sources before
                                                 pushing new ones; use --lagoon-clear-problems=false to disable (default true)
      --lagoon-environment string                The Lagoon environment name (env: LAGOON_ENVIRONMENT)
      --lagoon-insights-remote-base-url string   Insights Remote base url, used for deleting problems
                                                  (default "http://lagoon-remote-insights-remote.lagoon.svc")
      --lagoon-insights-remote-endpoint string   Insights Remote Problems endpoint
                                                  (default "http://lagoon-remote-insights-remote.lagoon.svc/problems")
      --lagoon-project string                    The Lagoon project name (env: LAGOON_PROJECT)
      --lagoon-push-problems-to-insights         Push audit facts to Lagoon via Insights Remote
      --lagoon-source string                     Source to use for Problems pushed to Lagoon (default "Shipshape")
      --output-file string                       Path to output file
                                                 (env: SHIPSHAPE_OUTPUT_FILE)
      --output-file-format string                Format for the output file [pretty|table|json|junit]
                                                 (env: SHIPSHAPE_OUTPUT_FILE_FORMAT)
  -o, --output-format string                     Output format [pretty|table|json|junit]
                                                 (env: SHIPSHAPE_OUTPUT_FORMAT) (default "pretty")
  -r, --remediate                                Run remediation for supported checks

Global Flags:
  -d, --debug              Display debug information - equivalent to --log-level debug
  -x, --exclude-db         Exclude checks requiring a database; overrides
                           any db checks specified by '--types'
  -f, --file strings       Path to the file containing the checks.
                           Can be specified as comma-separated single argument or
                           using --file (-f) multiple times (default [shipshape.yml])
  -l, --log-level string   Level of logs to display (default "warn")
  -t, --types strings      List of checks to run; default is empty, which will
                           run all checks. Can be specified as comma-separated
                           single argument or using --types (-t) multiple times
  -v, --verbose            Display verbose output - equivalent to --log-level info
```

### Merging multiple config files

You can pass more than one config file and Shipshape will merge them into a
single configuration before running:

```sh
shipshape run -f base.yml -f overrides.yml .
```

Files are merged in the order given on the command line — the first file is the
base, and each subsequent file is layered on top. Merging follows these rules:

- **Maps are merged recursively.** A later file can set or override an
  individual field of a plugin instance without repeating the whole block.
- **Later files win.** When the same scalar value is defined in more than one
  file, the value from the last file on the command line is used.
- **Lists are replaced, not combined.** When a later file defines a list value
  (for example `skip-dir` or `exclude-pattern`), it wholly replaces the earlier
  list rather than appending to it.

Every override is logged at `warn` level with the config key that was changed,
so you can audit what a later file altered.

Merging applies to v2 config only. All files in a single run must be v2 config;
mixing v1 (`checks:`) and v2 (`collect:`) config files in one invocation is not
supported and will return an error.

#### Example

`base.yml`:
```yaml
collect:
  sensitive-files:
    file:lookup:
      path: web/sites/default/files
      pattern: '.*\.(sql|php)?$'
      skip-dir:
        - private
```

`overrides.yml` — narrows the path and replaces the skip list, leaving the
pattern untouched:
```yaml
collect:
  sensitive-files:
    file:lookup:
      path: web/sites/default/files/public
      skip-dir:
        - tmp
```

The merged result keeps `pattern` from `base.yml`, takes `path` from
`overrides.yml`, and uses the `skip-dir` list from `overrides.yml`
(`[tmp]`, not `[private, tmp]`).

### Pushing problems to Lagoon

Pushing and deleting Lagoon Problems uses insights-remote's token-authenticated
HTTP endpoints exclusively — the Lagoon GraphQL API is not used for this, so no
`project`/`environment` lookup or API credentials are required. Only
`--lagoon-insights-remote-endpoint` (or `--lagoon-insights-remote-base-url` for
deletion) and the in-cluster bearer token at
`/var/run/secrets/lagoon/dynamic/insights-token/INSIGHTS_TOKEN` are needed.

By default (`--lagoon-clear-problems=true`), Shipshape deletes existing Lagoon
problems for its owned sources on every run, before pushing new ones — not only
when a run has zero breaches. Set `--lagoon-clear-problems=false` to disable
deletion entirely.

The Problem `source` can be overridden per-analyser via the analyser's
`source:` field, falling back to the plugin-level `--lagoon-source` /
`source:` when unset, and to `"Shipshape"` if neither is set. See
[Analysing data](analyse) for the per-analyser field.

## Next steps

  - [Connections](connections)
  - [Collecting data](collect)
  - [Analysing data](analyse)
  - [Remediating breaches](remediate)
  - [Outputs](outputs)
