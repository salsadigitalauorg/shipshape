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
    not-empty:
      description: 'Disallowed php scripts found'
      input: disallowed-php-scripts
      severity: high
  sensitive-public-files-found:
    not-empty:
      description: 'Sensitive files found in public directory'
      input: sensitive-public-files
      severity: high
```

Execute the policy:
```sh
shipshape run .
```

See the [configuration](/config) documentation for more information.

```
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
  -o, --output string                            Output format [json|junit|simple|table]
                                                 (env: SHIPSHAPE_OUTPUT_FORMAT) (default "simple")
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

