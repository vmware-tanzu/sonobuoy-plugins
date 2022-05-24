FROM debian:stretch-slim

# Install kubectl
# Note: Latest version may be found on:
# https://aur.archlinux.org/packages/kubectl-bin/
ADD https://storage.googleapis.com/kubernetes-release/release/v1.14.1/bin/linux/amd64/kubectl /usr/local/bin/kubectl

ENV HOME=/config

# Basic checks tools for ease of use.
RUN apt-get update && \
    apt-get -y install net-tools && \
    apt-get -y install curl && \
    apt-get install -y jq && \
    chmod +x /usr/local/bin/kubectl && \
    kubectl version --client

COPY ./run.sh ./run.sh
COPY ./sonobuoy /usr/local/bin/sonobuoy

RUN chmod +x ./run.sh

ENTRYPOINT ["./run.sh"]