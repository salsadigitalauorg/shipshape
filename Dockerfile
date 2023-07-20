FROM --platform=$BUILDPLATFORM golang:1.20 AS builder

ARG VERSION
ARG COMMIT

ADD . $GOPATH/src/github.com/salsadigitalauorg/shipshape/

WORKDIR $GOPATH/src/github.com/salsadigitalauorg/shipshape

ENV CGO_ENABLED 0

ARG TARGETOS TARGETARCH

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} && \
    go mod tidy && \
    go generate ./... && \
    go build -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o build/shipshape

FROM scratch

COPY --from=builder /go/src/github.com/salsadigitalauorg/shipshape/build/shipshape /usr/local/bin/shipshape
