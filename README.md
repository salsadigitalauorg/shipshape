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

### Security scanning

Dependabot (`.github/dependabot.yml`) handles dependency CVEs, including
indirect ones.
[`govulncheck`](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) covers
what it can't:

- **Standard library and toolchain vulnerabilities** — Dependabot's `gomod`
  ecosystem doesn't bump Go itself. Go is pinned in `go.mod`, both CI
  workflows, and the `Dockerfile`.
- **Reachability triage** — this module has 23 direct dependencies but over
  300 in the full graph. govulncheck reports whether our code actually calls
  the vulnerable symbol, so advisories deep in build tooling can be
  deprioritised instead of all treated as equally urgent.
- **Advisories with no fix yet** — Dependabot only opens a PR once a patched
  version exists; govulncheck reports regardless.

Install it once:
```sh
go install golang.org/x/vuln/cmd/govulncheck@latest
```

Scan the source. `go generate ./...` is required first — the generated
`*_gen.go` files are not committed, and govulncheck has to build the packages:
```sh
go generate ./...
govulncheck ./...
```

Output is split into `=== Symbol Results ===`, `=== Package Results ===` and
`=== Module Results ===`. Anything under Symbol Results has a call stack from
our code — fix these first. Module/Package Results without a call stack are
present in the dependency graph but not reachable from our code; lower
priority, but still worth tracking.

Useful variations:
```sh
govulncheck -show verbose ./...     # progress output + full detail on findings
govulncheck -scan module            # module-level only, fast, no build required
govulncheck -format json ./...      # machine-readable
govulncheck ./pkg/checks/...        # narrow to a subtree
```

Scan a built binary — checks what actually shipped, no source or
`go generate` needed:
```sh
go build -ldflags="-s -w" -o build/shipshape .
govulncheck -mode binary build/shipshape
```
Binary mode has no call stacks and reports at coarser granularity than a
source scan, since release binaries are stripped (`-ldflags="-s -w"`).

Exit code is `0` when no vulnerabilities are found and non-zero when there
are (except with `-format json`/`sarif`/`openvex`, which always exit `0` —
check the output instead).

To resolve a finding, upgrade to the fixed version named in the report, then
re-scan:
```sh
go get -u github.com/example/module@v1.2.3
go mod tidy
go generate ./... && govulncheck ./...
```
If no fixed version exists yet, check reachability with `-show verbose` and
record the decision in the PR.

### Documentation
```sh
cd docs
npm install
npm run dev
```
