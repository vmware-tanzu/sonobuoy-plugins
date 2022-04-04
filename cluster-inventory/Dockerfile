FROM golang:1.17-buster as build

COPY . /cluster-inventory
WORKDIR /cluster-inventory

RUN go build

FROM debian:buster-slim

COPY --from=build /cluster-inventory/cluster-inventory /cluster-inventory
CMD ["bash", "-c", "/cluster-inventory"]
