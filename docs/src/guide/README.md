# Quick-start

## Installation

### MacOS

The preferred method is installation via [Homebrew](https://brew.sh/).
```sh
brew install salsadigitalauorg/shipshape
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
docker run --rm ghcr.io/salsadigitalauorg/shipshape:latest shipshape --version
```

Or add to your docker image:
```Dockerfile
COPY --from=ghcr.io/salsadigitalauorg/shipshape:latest /usr/local/bin/shipshape /usr/local/bin/shipshape
```

## Usage
Create a config file. Can be as simple as:
```yaml
# shipshape.yml
checks:
  file:
    - name: Illegal files
      path: web
      disallowed-pattern: '^(adminer|phpmyadmin|bigdump)?\.php$'
```
See the [configuration](/config) documentation for more information.

```
$ shipshape -h
Shipshape

Run checks quickly on your project.

Usage:
  shipshape [dir]

Flags:
  -e, --error-code      Exit with error code if a failure is detected (env: SHIPSHAPE_ERROR_ON_FAILURE)
  -d, --exclude-db      Exclude checks requiring a database; overrides any db checks specified by '--types'
  -f, --file string     Path to the file containing the checks (default "shipshape.yml")
  -h, --help            Displays usage information
  -o, --output string   Output format [json|junit|simple|table] (env: SHIPSHAPE_OUTPUT_FORMAT) (default "simple")
  -t, --types strings   Comma-separated list of checks to run; default is empty, which will run all checks
  -v, --version         Displays the application version
```

