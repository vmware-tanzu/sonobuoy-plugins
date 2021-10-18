##########################################################################
# Copyright 2017 Heptio Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Get qemu-user-static
ARG IMAGEARCH
FROM alpine:3.9.2 as qemu
RUN apk add --no-cache curl
ARG QEMUVERSION=6.1.0-6
ARG QEMUARCH

SHELL ["/bin/ash", "-o", "pipefail", "-c"]

RUN curl -fsSL https://github.com/multiarch/qemu-user-static/releases/download/v${QEMUVERSION}/qemu-${QEMUARCH}-static.tar.gz | tar zxvf - -C /usr/bin
RUN chmod +x /usr/bin/qemu-*

FROM ${IMAGEARCH}buildpack-deps:stable-scm
MAINTAINER John Schnake "jschnake@vmware.com"

ARG QEMUARCH
COPY --from=qemu /usr/bin/qemu-${QEMUARCH}-static /usr/bin/

RUN apt-get update && apt-get -y --no-install-recommends install \
    ca-certificates \
    && rm -rf /var/cache/apt/* \
    && rm -rf /var/lib/apt/lists/*
COPY get_systemd_logs.sh /get_systemd_logs.sh

RUN rm -f /usr/bin/qemu-${QEMUARCH}-static

CMD ["/bin/bash", "-c", "/get_systemd_logs.sh"]
