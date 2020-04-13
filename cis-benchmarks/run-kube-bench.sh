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
        "gke")
            # Although we support GKE as a custom distribution, it uses the default configuration.
            ;;
        "")
            # If unset, use default config file.
            ;;
        *)
            ;;
    esac

    echo $config
}

# Returns 0 if the $1 is less than or equal to $2, otherwise 1
verlte() {
    [  "$1" = "`printf "%s\n%s" $1 $2 | sort -V | head -n1`" ]
}

# Returns 0 if the $1 is less than $2, otherwise 1
verlt() {
    [ "$1" = "$2" ] && return 1 || verlte "$1" "$2"
}

# Return either the version or benchmark flag to be used.
# If a distribution requiring a particular benchmark is provided, this will be returned,
# otherwise return the version flag with the version if specified. This enables users
# to still use the version auto-detect feature if desired.
get_version_or_benchmark_flag() {
    local vb_flag=""

    # If the distribution is GKE, then we may need to explicitly set the benchmark version that should be used.
    case $DISTRIBUTION in
        "gke")
            # The GKE specific benchmark is only suitable for Kubernetes 1.15 and later. If the provided
            # version is less than this, fall back to specifying the version manually.
            if [ -n "$KUBERNETES_VERSION" ] && verlt "$KUBERNETES_VERSION" "1.15" ; then
                vb_flag="--version $KUBERNETES_VERSION"
            else
                vb_flag="--benchmark gke-1.0"
            fi
            ;;
        *)
            if [ -n "$KUBERNETES_VERSION" ]; then
                vb_flag="--version $KUBERNETES_VERSION"
            fi
            ;;
    esac


    echo $vb_flag
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

    # This target is only compatible when running the GKE benchmark.
    if [ "$TARGET_MANAGED_SERVICES" = true ]; then
        targets="${targets} managedservices"
    fi

    echo $targets
}

run_kube_bench() {
    local config="$(get_config)"
    local vb_flag="$(get_version_or_benchmark_flag)"
    local targets="$(get_targets)"

    for target in $targets; do
        kube-bench --config $config run $vb_flag --targets $target --outputfile /tmp/results/$target.xml --junit
    done

    tar czf /tmp/results/results.tar.gz /tmp/results/*.xml
    echo -n /tmp/results/results.tar.gz > /tmp/results/done
}

run_kube_bench
