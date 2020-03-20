#!/bin/sh

##########################################################################
# Copyright the Sonobuoy contributors 2020
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit

# Return the config file to be used by kube-bench.
# If the specified distribution is supported, the path to a custom
# configuration file will be returned. If not, the default configuration
# is used.
get_config() {
    # Assume default config path of cfg/config.yaml which is relative
    # to the workdir set in base image.
    local config="cfg/config.yaml"

    case $DISTRIBUTION in
        # We will add custom configurations for different distributions here.
        "entpks")
            config="cfg/entpks.yaml"
            ;;
        "")
            # If unset, use default config file.
            ;;
        *)
            ;;
    esac

    echo $config
}

# Return the version flag with the version if specified. This enables users
# to still use the version auto-detect feature if needed.
get_version_flag() {
    local version_flag=""

    if [ -n "$KUBERNETES_VERSION" ]; then
        version_flag="--version $KUBERNETES_VERSION"
    fi

    echo $version_flag
}

# Return a space separated list of targets to provide to kube-bench.
get_targets() {
    local targets

    if [ "$TARGET_MASTER" = true ]; then
        targets="${targets} master"
    fi

    if [ "$TARGET_NODE" = true ]; then
        targets="${targets} node"
    fi

    # Other targets are only compatible with kube-bench for Kubernetes 1.15 and later.
    # We could prevent them from being added, however we may not always know the
    # version being tested as the user may be relying on version being auto-detected.
    if [ "$TARGET_CONTROLPLANE" = true ]; then
        targets="${targets} controlplane"
    fi

    if [ "$TARGET_ETCD" = true ]; then
        targets="${targets} etcd"
    fi

    if [ "$TARGET_POLICIES" = true ]; then
        targets="${targets} policies"
    fi

    echo $targets
}

run_kube_bench() {
    local config="$(get_config)"
    local version_flag="$(get_version_flag)"
    local targets="$(get_targets)"

    for target in $targets; do
        kube-bench --config $config run $version_flag --targets $target --outputfile /tmp/results/$target.xml --junit
    done

    tar czf /tmp/results/results.tar.gz /tmp/results/*.xml
    echo -n /tmp/results/results.tar.gz > /tmp/results/done
}

run_kube_bench
