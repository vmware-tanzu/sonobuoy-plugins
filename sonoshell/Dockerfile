FROM golang:1.17-buster as build

run mkdir -p /_sonoshell
COPY ./pkg /_sonoshell/pkg
COPY ./go.sum /_sonoshell/go.sum
COPY ./go.mod /_sonoshell/go.mod

WORKDIR /_sonoshell

# Needed to build successfully
RUN go mod tidy

RUN go build -o sonoshell ./pkg

FROM debian:buster-slim

COPY --from=build /_sonoshell/sonoshell /sonoshell
COPY ./test.yaml /test.yaml

CMD ["/sonoshell", "/test.yaml"]
