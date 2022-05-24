FROM golang:1.17.10 AS build

WORKDIR /go/src/github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner

COPY go.mod go.sum ./
COPY cmd cmd
COPY internal internal
COPY api api

ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o /go/bin/ ./cmd/reliability-scanner/

FROM gcr.io/distroless/base AS final
COPY --from=build /go/bin/reliability-scanner /bin/reliability-scanner
CMD ["/bin/reliability-scanner","scan"]
