FROM golang:1.17-buster as build

WORKDIR /src
COPY go.* /src
ENV CGO_ENABLED=0
RUN go mod download

COPY main.go /src
COPY ./lib /src/lib
COPY ./cmd /src/cmd

RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux GOARCH=amd64 go generate ./... && go build -o sonolark .

FROM gcr.io/distroless/static:nonroot as dist

COPY --from=build /src/sonolark .
CMD ["./sonolark"]
