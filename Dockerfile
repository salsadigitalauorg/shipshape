FROM golang:1.17 AS builder

ADD . $GOPATH/src/github.com/salsadigitalauorg/shipshape/

WORKDIR $GOPATH/src/github.com/salsadigitalauorg/shipshape

ENV CGO_ENABLED 0

RUN go build -ldflags="-s -w" -o build/shipshape

FROM scratch

COPY --from=builder /go/src/github.com/salsadigitalauorg/shipshape/build/shipshape /usr/local/bin/shipshape
