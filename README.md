# ShipShape
[![Go](https://github.com/salsadigitalauorg/shipshape/actions/workflows/go.yml/badge.svg)](https://github.com/salsadigitalauorg/shipshape/actions/workflows/go.yml)

## Build
```sh
go build -ldflags="-s -w" -o build/shipshape .
```

## Run tests
```sh
go test -v ./... -coverprofile=build/coverage.out
```

View coverage results:
```sh
go tool cover -html=build/coverage.out
```

## Documentation
```sh
cd docs
npm install
npm run dev
```
