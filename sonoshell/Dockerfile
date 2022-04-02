FROM golang:1.17-buster as build

RUN mkdir -p /_sonoshell
COPY ./go.sum /_sonoshell/go.sum
COPY ./go.mod /_sonoshell/go.mod
WORKDIR /_sonoshell
RUN go mod download
COPY ./pkg /_sonoshell/pkg

RUN go build -o sonoshell ./pkg

FROM debian:buster-slim

COPY --from=build /_sonoshell/sonoshell /sonoshell
COPY ./.test.yaml /test.yaml

CMD ["/sonoshell", "/test.yaml"]
