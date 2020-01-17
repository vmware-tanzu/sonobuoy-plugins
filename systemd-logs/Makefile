# Copyright 2017 Heptio Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Note the only reason we are creating this is because upstream
# does not yet publish a released e2e container
# https://github.com/kubernetes/kubernetes/issues/47920

TARGET = systemd-logs
REGISTRY ?= sonobuoy
IMAGE = $(REGISTRY)/$(TARGET)
DOCKER ?= docker
DIR := ${CURDIR}
VERSION ?= v0.2

ARCH    ?= amd64
LINUX_ARCHS = amd64 arm64 ppc64le
PLATFORMS = linux/amd64,linux/arm64,linux/ppc64le

IMAGEARCH ?=
QEMUARCH  ?=

MANIFEST_TOOL_VERSION := v1.0.0

ifneq ($(ARCH),amd64)
TARGET = systemd-logs-$(ARCH)
endif

ifeq ($(ARCH),amd64)
IMAGEARCH =
QEMUARCH  = x86_64
else ifeq ($(ARCH),arm)
IMAGEARCH = arm32v7/
QEMUARCH  = arm
else ifeq ($(ARCH),arm64)
IMAGEARCH = arm64v8/
QEMUARCH  = aarch64
else ifeq ($(ARCH),ppc64le)
IMAGEARCH = ppc64le/
QEMUARCH  = ppc64le
else
$(error unknown arch "$(ARCH)")
endif

all: container

pre-cross:
	docker run --rm --privileged multiarch/qemu-user-static:register --reset

build-container: get_systemd_logs.sh
	$(DOCKER) build --build-arg IMAGEARCH=$(IMAGEARCH) \
               --build-arg QEMUARCH=$(QEMUARCH) \
               -t $(REGISTRY)/$(TARGET):latest -t $(REGISTRY)/$(TARGET):$(VERSION) .

container: pre-cross
	for arch in $(LINUX_ARCHS); do \
		$(MAKE) build-container ARCH=$$arch TARGET="systemd-logs-$$arch"; \
	done

push-image:
	$(DOCKER) push $(REGISTRY)/$(TARGET):latest
	$(DOCKER) push $(REGISTRY)/$(TARGET):$(VERSION)

push_manifest:
	./manifest-tool push from-args --platforms $(PLATFORMS) --template $(REGISTRY)/$(TARGET)-ARCH:$(VERSION) --target $(REGISTRY)/$(TARGET):$(VERSION)

pre-push:
	curl -sSL https://github.com/estesp/manifest-tool/releases/download/$(MANIFEST_TOOL_VERSION)/manifest-tool-linux-amd64 > manifest-tool
	chmod +x manifest-tool

push: pre-push container
	for arch in $(LINUX_ARCHS); do \
		$(MAKE) push-image TARGET="systemd-logs-$$arch"; \
	done

	$(MAKE) push_manifest VERSION=latest
	$(MAKE) push_manifest

.PHONY: all container push

clean-image:
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):latest $(REGISTRY)/$(TARGET):latest || true
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):latest $(REGISTRY)/$(TARGET):$(VERSION) || true
clean:
	rm -f manifest-tool*
	for arch in $(LINUX_ARCHS); do \
		$(MAKE) clean-image TARGET=$(TARGET)-$$arch; \
	done
