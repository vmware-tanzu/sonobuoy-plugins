FROM golang:1.17-buster as build

COPY . /who-can
WORKDIR /who-can

RUN go build

FROM debian:buster-slim

COPY --from=build /who-can/who-can /who-can
CMD ["bash", "-c", "/who-can"]
