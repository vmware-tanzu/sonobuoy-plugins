#!/bin/sh

# This doesn't fully generate the plugins but handles the bulk of the work.
# See the README.md for more details.
set +x

VERSIONS=("image-repo-list-2004" "image-repo-list-master" "image-repo-list")
FOCUS='\[Conformance\]|\[NodeConformance\]|\[sig-windows\]|\[sig-apps\].CronJob|\[sig-api-machinery\].ResourceQuota|\[sig-scheduling\].SchedulerPreemption|\[sig-autoscaling\].\[Feature:HPA\]'
SKIP='\[LinuxOnly\]|\[Serial\]|GMSA|Guestbook.application.should.create.and.stop.a.working.application'
rm image-repo-list*

for REPO_VERSION in "${VERSIONS[@]}" 
do
curl --show-error --silent "https://raw.githubusercontent.com/kubernetes-sigs/windows-testing/master/images/${REPO_VERSION}" -o ${REPO_VERSION}

sonobuoy gen plugin e2e \
--e2e-focus=${FOCUS} \
--e2e-skip=${SKIP} \
--configmap=./${REPO_VERSION} \
| yq e '(.spec.env.[] | select(.name == "E2E_EXTRA_ARGS") | .value ) = "--progress-report-url=http://localhost:8099/progress --node-os-distro=windows"' - \
| yq e '.spec.image = "k8s.gcr.io/conformance:$SONOBUOY_K8S_VERSION"' - > win-e2e-${REPO_VERSION}.yaml

# Add
#    - name: KUBE_TEST_REPO_LIST
#      value: /tmp/sonobuoy/config/image-repo-list-XYZ
# to each plugin env vars. Should be able to do this with yq but struggling to find the syntax.

done
