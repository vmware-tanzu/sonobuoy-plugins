FROM golang:1.17-buster as build

# Install kubectl
# Note: Latest version may be found on:
# https://aur.archlinux.org/packages/kubectl-bin/
RUN wget https://storage.googleapis.com/kubernetes-release/release/v1.21.3/bin/linux/amd64/kubectl -O /usr/bin/kubectl && \
    chmod +x /usr/bin/kubectl

# Copy go.[mod|sum] first and download deps to utilize docker cache.
COPY go.sum /requirements-check/go.sum
COPY go.mod /requirements-check/go.mod
WORKDIR /requirements-check
RUN go mod download

# Copy the rest of the project files.
COPY ./pkg /requirements-check/pkg

RUN go build -o binary ./pkg

FROM debian:buster-slim
# Add jq; moving just the binary caused issues with some dynamic libraries.
RUN apt-get update && \
    apt-get install -y jq

COPY --from=build /requirements-check/binary /requirements-check
COPY --from=build /usr/bin/kubectl /usr/bin/kubectl

CMD ["bash", "-c", "/requirements-check"]
