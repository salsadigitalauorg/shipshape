# ShipShape
[![Go](https://github.com/salsadigitalauorg/shipshape/actions/workflows/go.yml/badge.svg)](https://github.com/salsadigitalauorg/shipshape/actions/workflows/go.yml)
[![Docker](https://github.com/salsadigitalauorg/shipshape/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/salsadigitalauorg/shipshape/actions/workflows/docker-publish.yml)
[![Coverage Status](https://coveralls.io/repos/github/salsadigitalauorg/shipshape/badge.svg?branch=main)](https://coveralls.io/github/salsadigitalauorg/shipshape?branch=main)

## Installation

### Binary

  - Download the binary for your OS (Linux/MacOS) and platform (amd64/arm64) from the [latest release](https://github.com/salsadigitalauorg/shipshape/releases/latest)
  - Untar and move shipshape to /usr/local/bin/shipshape

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
  -f, --file string     Path to the file containing the checks (default "shipshape.yml")
  -h, --help            Displays this help
  -o, --output string   Output format (json|junit|simple|table) (default "simple")
  -t, --types strings   Comma-separated list of checks to run; default is empty, which will run all checks
  -v, --version         Displays the application version
```

## Documentation
Check out our documentation at https://salsadigitalauorg.github.io/shipshape/. Keep in mind that this is still a work in progress, so please go easy.

## Local development

### Build
```sh
git clone git@github.com:salsadigitalauorg/shipshape.git && cd shipshape
go build -ldflags="-s -w" -o build/shipshape .
go run . -h
```

### Run tests
```sh
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
