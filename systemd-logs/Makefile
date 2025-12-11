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

REGISTRY ?= sonobuoy
TARGET = systemd-logs
IMAGE = $(REGISTRY)/$(TARGET)

DOCKER ?= docker
DIR := ${CURDIR}
VERSION ?= v0.5
PLATFORMS = linux/amd64,linux/arm64,linux/ppc64le,linux/s390x,linux/riscv64

# for container load only
ARCH ?= amd64

all: container

container: get_systemd_logs.sh
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):latest \
		.

load-container: get_systemd_logs.sh
	echo "start load contianer for $(ARCH)"
	docker buildx build \
		--platform linux/$(ARCH) \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):latest \
		--load \
		.
	echo "done! load contianer for $(ARCH)"

push: container
	# Will trigger the caching mechanism
	docker buildx build \
			--platform $(PLATFORMS) \
			-t $(IMAGE):$(VERSION) \
			-t $(IMAGE):latest \
			--push \
			.

.PHONY: all container load-container push

clean-image:
	$(DOCKER) rmi $(IMAGE):latest $(IMAGE):$(VERSION) || true

clean: clean-image
