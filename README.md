# ShipShape
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/salsadigitalauorg/shipshape)
[![Go Report Card](https://goreportcard.com/badge/github.com/salsadigitalauorg/shipshape)](https://goreportcard.com/report/github.com/salsadigitalauorg/shipshape)
[![Coverage Status](https://coveralls.io/repos/github/salsadigitalauorg/shipshape/badge.svg?branch=1.x)](https://coveralls.io/github/salsadigitalauorg/shipshape?branch=1.x)
[![Release](https://img.shields.io/github/v/release/salsadigitalauorg/shipshape)](https://github.com/salsadigitalauorg/shipshape/releases/latest)

## Config versions

Shipshape has two configuration formats, and the same binary runs both.

- **1.x — recommended.** The composable pipeline (`collect` / `analyse` /
  `output`) is the format to use going forward. It's used in production and is
  where new capabilities land.
- **0.x — fully supported.** The legacy `checks:` format continues to run
  without changes. There is no forced migration — migrate when it suits you.

1.x takes a compose-don't-port approach: instead of one dedicated check per
task, you wire a small set of general-purpose plugins together to get the same
outcome.

> **Closing the gaps:** a few 0.x checks don't have a documented 1.x recipe
> yet. We're actively closing that gap. See the
> [gaps matrix](https://salsadigitalauorg.github.io/shipshape/1.x/guide/gaps.html)
> for what's covered today and the
> [roadmap](https://salsadigitalauorg.github.io/shipshape/1.x/guide/roadmap.html)
> for what's coming. Full comparison:
> [Config versions](https://salsadigitalauorg.github.io/shipshape/1.x/guide/versions.html).

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

## Documentation
Check out our documentation at https://salsadigitalauorg.github.io/shipshape/.

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
