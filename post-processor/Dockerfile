FROM golang:1.17-buster as build

# Install kubectl
# Note: Latest version may be found on:
# https://aur.archlinux.org/packages/kubectl-bin/
RUN wget https://storage.googleapis.com/kubernetes-release/release/v1.21.3/bin/linux/amd64/kubectl -O /usr/bin/kubectl && \
    chmod +x /usr/bin/kubectl

RUN wget https://github.com/vmware-tanzu/carvel-ytt/releases/download/v0.40.1/ytt-linux-amd64 -O /usr/bin/ytt && \
    chmod +x /usr/bin/ytt

COPY go.sum /src/go.sum
COPY go.mod /src/go.mod
WORKDIR /src
RUN go mod download

COPY ./cmd /src/cmd
COPY main.go /src/main.go
RUN go build -o binary

FROM debian:buster-slim

COPY --from=build /src/binary /sonobuoy-processor
COPY --from=build /usr/bin/kubectl /usr/bin/kubectl
COPY --from=build /usr/bin/ytt /usr/bin/ytt
RUN chmod +x /sonobuoy-processor

# Add jq; moving just the binary caused issues with some dynamic libraries.
RUN apt-get update && \
    apt-get install -y jq

CMD ["/bin/sh", "-c", "/sonobuoy-processor"]
