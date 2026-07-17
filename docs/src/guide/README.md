# Quick-start

## Installation

### MacOS

The preferred method is installation via [Homebrew](https://brew.sh/).
```sh
brew install salsadigitalauorg/shipshape/shipshape
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
      --lagoon-api-base-url string               Base url for the Lagoon API when pushing
                                                 problems to API (env: LAGOON_API_BASE_URL)
      --lagoon-api-token string                  Lagoon API token when pushing problems
                                                 to API (env: LAGOON_API_TOKEN)
      --lagoon-insights-remote-endpoint string   Insights Remote Problems endpoint
                                                  (default "http://lagoon-remote-insights-remote.lagoon.svc/problems")
      --lagoon-push-problems-to-insights         Push audit facts to Lagoon via Insights Remote
  -o, --output-format string                     Output format for stdout [pretty|table|json|junit]
                                                 (overrides config file)
      --output-file string                       File to write output to
      --output-file-format string                Format for file output [pretty|table|json|junit]
                                                 (defaults to stdout format)
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

## Next steps

  - [Connections](connections)
  - [Collecting data](collect)
  - [Analysing data](analyse)
  - [Remediating breaches](remediate)
  - [Outputs](outputs)
