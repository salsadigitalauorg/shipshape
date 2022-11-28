# ShipShape
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/salsadigitalauorg/shipshape)
[![Go Report Card](https://goreportcard.com/badge/github.com/salsadigitalauorg/shipshape)](https://goreportcard.com/report/github.com/salsadigitalauorg/shipshape)
[![Coverage Status](https://coveralls.io/repos/github/salsadigitalauorg/shipshape/badge.svg?branch=main)](https://coveralls.io/github/salsadigitalauorg/shipshape?branch=main)
[![Release](https://img.shields.io/github/v/release/salsadigitalauorg/shipshape)](https://github.com/salsadigitalauorg/shipshape/releases/latest)

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
See the [configuration](https://salsadigitalauorg.github.io/shipshape/config) documentation for more information.

```
$ shipshape -h
Shipshape

Run checks quickly on your project.

Usage:
  shipshape [dir]

Flags:
      --dump-config     Dump the final config - useful to make sure multiple config files are being merged as expected
  -e, --error-code      Exit with error code if a failure is detected (env: SHIPSHAPE_ERROR_ON_FAILURE)
  -d, --exclude-db      Exclude checks requiring a database; overrides any db checks specified by '--types'
  -f, --file strings    Path to the file containing the checks. Can be specified as comma-separated single argument or using --types multiple times (default [shipshape.yml])
  -h, --help            Displays usage information
      --list-checks     List available checks
  -o, --output string   Output format [json|junit|simple|table] (env: SHIPSHAPE_OUTPUT_FORMAT) (default "simple")
  -t, --types strings   List of checks to run; default is empty, which will run all checks. Can be specified as comma-separated single argument or using --types multiple times
  -v, --version         Displays the application version
```

## Documentation
Check out our documentation at https://salsadigitalauorg.github.io/shipshape/. Keep in mind that this is still a work in progress, so please go easy.

## Local development

### Build
```sh
git clone git@github.com:salsadigitalauorg/shipshape.git && cd shipshape
go generate ./...
go build -ldflags="-s -w" -o build/shipshape .
go run . -h
```

### Run tests
```sh
go generate ./...
go test -v ./... -coverprofile=build/coverage.out
```

View coverage results:
```sh
go tool cover -html=build/coverage.out
```

### Documentation
```sh
cd docs
npm install
npm run dev
```
