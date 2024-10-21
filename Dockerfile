FROM --platform=$BUILDPLATFORM golang:1.21 AS builder

ARG VERSION
ARG COMMIT

ADD . $GOPATH/src/github.com/salsadigitalauorg/shipshape/

WORKDIR $GOPATH/src/github.com/salsadigitalauorg/shipshape

ENV CGO_ENABLED=0

ARG TARGETOS TARGETARCH

RUN go mod tidy && \
    go generate ./... && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
      -ldflags="-s -w \
        -X github.com/salsadigitalauorg/shipshape/cmd.version=${VERSION} \
        -X github.com/salsadigitalauorg/shipshape/cmd.commit=${COMMIT}" \
      -o build/shipshape

FROM scratch

COPY --from=builder /go/src/github.com/salsadigitalauorg/shipshape/build/shipshape /usr/local/bin/shipshape
