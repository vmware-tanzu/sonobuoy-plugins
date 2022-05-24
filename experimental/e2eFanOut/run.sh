#!/usr/bin/env bash

set -x

# This is the entrypoint for the image and meant to wrap the
# logic of gathering/reporting results to the Sonobuoy worker.

results_dir="${RESULTS_DIR:-/tmp/sonobuoy/results}"

# Number of buckets indicate the fraction of tests to run at one time. Each sub-plugin
# will always have 5 tests, but # buckets determines how many of those plugins to run at once.
# More buckets mean fewer tests at once which is slower but puts less load on server.
num_buckets="${NUM_BUCKETS:-5}"
stop_at_bucket="${STOP_AFTER_BUCKET:-$num_buckets}"

mkdir -p ${results_dir}

# saveResults prepares the results for handoff to the Sonobuoy worker.
# See: https://github.com/vmware-tanzu/sonobuoy/blob/main/site/content/docs/main/plugins.md
saveResults() {
  cd ${results_dir}

  # Sonobuoy worker expects a tar file.
	tar czf results.tar.gz *

	# Signal to the worker that we are done and where to find the results.
	printf ${results_dir}/results.tar.gz > ${results_dir}/done
}

# Ensure that we tell the Sonobuoy worker we are done regardless of results.
trap saveResults EXIT

# Make the plugins
rm -r tmpPlugins
mkdir tmpPlugins

# This is the list for conformance-lite (e.g. fast tests).
# Making 5 test per plugin.
sonobuoy e2e -f "\[Conformance\]" -s "Serial|Slow|Disruptive|\[sig-apps\] StatefulSet Basic StatefulSet functionality\[StatefulSetBasic\] should have a working scale subresource \[Conformance\]|\[sig-network\]EndpointSlice should create Endpoints and EndpointSlices for Pods matching aService \[Conformance\]|\[sig-api-machinery\] CustomResourcePublishOpenAPI \[Privileged:ClusterAdmin\]works for multiple CRDs of same group and version but different kinds \[Conformance\]|\[sig-auth\]ServiceAccounts ServiceAccountIssuerDiscovery should support OIDC discoveryof service account issuer \[Conformance\]|\[sig-network\] DNS should provideDNS for services  \[Conformance\]|\[sig-network\] DNS should resolve DNS ofpartial qualified names for services \[LinuxOnly\] \[Conformance\]|\[sig-apps\]Job should delete a job \[Conformance\]|\[sig-network\] DNS should provide DNSfor ExternalName services \[Conformance\]|\[sig-node\] Variable Expansion shouldsucceed in writing subpaths in container \[Slow\] \[Conformance\]|\[sig-apps\]Daemon set \[Serial\] should rollback without unnecessary restarts \[Conformance\]|\[sig-api-machinery\]Garbage collector should orphan pods created by rc if delete options say so\[Conformance\]|\[sig-network\] Services should have session affinity timeoutwork for service with type clusterIP \[LinuxOnly\] \[Conformance\]|\[sig-network\]Services should have session affinity timeout work for NodePort service \[LinuxOnly\]\[Conformance\]|\[sig-node\] InitContainer \[NodeConformance\] should not startapp containers if init containers fail on a RestartAlways pod \[Conformance\]|\[sig-apps\]Daemon set \[Serial\] should update pod when spec was updated and update strategyis RollingUpdate \[Conformance\]|\[sig-api-machinery\] CustomResourcePublishOpenAPI\[Privileged:ClusterAdmin\] works for multiple CRDs of same group but differentversions \[Conformance\]|\[sig-apps\] StatefulSet Basic StatefulSet functionality\[StatefulSetBasic\] Burst scaling should run to completion even with unhealthypods \[Slow\] \[Conformance\]|\[sig-node\] Probing container should be restartedwith a exec .cat /tmp/health. liveness probe \[NodeConformance\] \[Conformance\]|\[sig-network\]Services should be able to switch session affinity for service with type clusterIP\[LinuxOnly\] \[Conformance\]|\[sig-node\] Probing container with readinessprobe that fails should never be ready and never restart \[NodeConformance\]\[Conformance\]|\[sig-api-machinery\] Watchers should observe add, update, anddelete watch notifications on configmaps \[Conformance\]|\[sig-scheduling\]SchedulerPreemption \[Serial\] PriorityClass endpoints verify PriorityClassendpoints can be operated with different HTTP methods \[Conformance\]|\[sig-api-machinery\]CustomResourceDefinition resources \[Privileged:ClusterAdmin\] Simple CustomResourceDefinitionlisting custom resource definition objects works  \[Conformance\]|\[sig-api-machinery\]CustomResourceDefinition Watch \[Privileged:ClusterAdmin\] CustomResourceDefinitionWatch watch on custom resource definition objects \[Conformance\]|\[sig-scheduling\]SchedulerPreemption \[Serial\] validates basic preemption works \[Conformance\]|\[sig-storage\]ConfigMap optional updates should be reflected in volume \[NodeConformance\]\[Conformance\]|\[sig-apps\] StatefulSet Basic StatefulSet functionality \[StatefulSetBasic\]Scaling should happen in predictable order and halt if any stateful pod is unhealthy\[Slow\] \[Conformance\]|\[sig-storage\] EmptyDir wrapper volumes should notcause race condition when used for configmaps \[Serial\] \[Conformance\]|\[sig-scheduling\]SchedulerPreemption \[Serial\] validates lower priority pod preemption by criticalpod \[Conformance\]|\[sig-storage\] Projected secret optional updates shouldbe reflected in volume \[NodeConformance\] \[Conformance\]|\[sig-apps\] CronJobshould schedule multiple jobs concurrently \[Conformance\]|\[sig-apps\] CronJobshould replace jobs when ReplaceConcurrent \[Conformance\]|\[sig-scheduling\]SchedulerPreemption \[Serial\] PreemptionExecutionPath runs ReplicaSets to verifypreemption running path \[Conformance\]|\[sig-apps\] StatefulSet Basic StatefulSetfunctionality \[StatefulSetBasic\] should perform canary updates and phasedrolling updates of template modifications \[Conformance\]|\[sig-apps\] StatefulSetBasic StatefulSet functionality \[StatefulSetBasic\] should perform rollingupdates and roll backs of template modifications \[Conformance\]|\[sig-node\]Probing container should have monotonically increasing restart count \[NodeConformance\]\[Conformance\]|\[sig-node\] Variable Expansion should verify that a failingsubpath expansion can be modified during the lifecycle of a container \[Slow\]\[Conformance\]|\[sig-node\] Probing container should \*not\* be restarted witha exec .cat /tmp/health. liveness probe \[NodeConformance\] \[Conformance\]|\[sig-node\]Probing container should \*not\* be restarted with a tcp:8080 liveness probe\[NodeConformance\] \[Conformance\]|\[sig-node\] Probing container should \*not\*be restarted with a /healthz http liveness probe \[NodeConformance\] \[Conformance\]|\[sig-apps\]CronJob should not schedule jobs when suspended \[Slow\] \[Conformance\]|\[sig-scheduling\]SchedulerPredicates \[Serial\] validates that there exists conflict betweenpods with same hostPort and protocol but one using 0\.0\.0\.0 hostIP \[Conformance\]|\[sig-apps\]CronJob should not schedule new jobs when ForbidConcurrent \[Slow\] \[Conformance\]|\[k8s\.io\]Probing container should \*not\* be restarted with a exec .cat /tmp/health.liveness probe \[NodeConformance\] \[Conformance\]|\[sig-apps\] StatefulSet\[k8s\.io\] Basic StatefulSet functionality \[StatefulSetBasic\] should performcanary updates and phased rolling updates of template modifications \[Conformance\]|\[sig-storage\]ConfigMap updates should be reflected in volume \[NodeConformance\] \[Conformance\]|\[sig-network\]Services should be able to switch session affinity for NodePort service \[LinuxOnly\]\[Conformance\]|\[k8s\.io\] Probing container with readiness probe that failsshould never be ready and never restart \[NodeConformance\] \[Conformance\]|\[sig-storage\]Projected configMap optional updates should be reflected in volume \[NodeConformance\]\[Conformance\]|\[k8s\.io\] Probing container should be restarted with a exec.cat /tmp/health. liveness probe \[NodeConformance\] \[Conformance\]|\[sig-api-machinery\]Garbage collector should delete RS created by deployment when not orphaning\[Conformance\]|\[sig-api-machinery\] Garbage collector should delete pods createdby rc when not orphaning \[Conformance\]|\[k8s\.io\] Probing container shouldhave monotonically increasing restart count \[NodeConformance\] \[Conformance\]|\[k8s\.io\]Probing container should \*not\* be restarted with a tcp:8080 liveness probe\[NodeConformance\] \[Conformance\]|\[sig-api-machinery\] Garbage collectorshould keep the rc around until all its pods are deleted if the deleteOptionssays so \[Conformance\]|\[sig-apps\] StatefulSet \[k8s\.io\] Basic StatefulSetfunctionality \[StatefulSetBasic\] should perform rolling updates and roll backsof template modifications \[Conformance\]" > tmptestlist
cat tmptestlist| tr '\n' '\0' | sed 's/\"/\*/g' | sed 's/\[/\\\[/g' | sed 's/\]/\\\]/g' | xargs -0 -n5 bash -c 'echo "$1|$2|$3|$4|$5"' bash > focusList

for (( c=0; c<num_buckets; c++ ))
do
  mkdir ./tmpPlugins/bucket${c}
done

# Use postprocessor to avoid repeating thousands of skips for each plugin.
cat << EOF > prefixFile
config-map:
  ytt-transform.yaml: |
    #@ load("@ytt:overlay", "overlay")

    #@overlay/match by=overlay.all
    #@overlay/match-child-defaults expects="0+"
    ---
    items:
      #@overlay/match by=overlay.all
      - items:
          #@overlay/match by=overlay.all
          - items:
            #@overlay/match by=overlay.subset({"status": "skipped"})
            #@overlay/remove
            -

    #@overlay/match by=overlay.all, expects="1+"
    ---
    items:
      #! file
      #@overlay/match by=overlay.all, expects="1+"
      -
        items:
        #! suite
        #@overlay/match by=overlay.subset({"name":"Kubernetes e2e suite"}), expects="0+"
        #@overlay/match by=lambda k,left,right: len(left["items"])==0, expects="0+"
        #@overlay/remove
        -

    #@overlay/match by=overlay.all, expects="1+"
    ---
    items:
      #! file
      #@overlay/match by=lambda k,left,right: len(left["items"])==0, expects="0+"
      #@overlay/remove
      -
podSpec:
  containers:
    - name: postprocessing
      image: schnake/postprocessor:v0
      command: ["/sonobuoy-processor"]
EOF

i=0
while read -r line; do
    #echo $line
    line="$(echo $line | sed 's/|*$//')"
    let bucket=i%num_buckets
    sonobuoy gen plugin e2e --e2e-focus="${line}" --e2e-skip= --plugin-env=e2e.E2E_PARALLEL=true --plugin-env=e2e.E2E_DRYRUN=true | sed "s/plugin-name: e2e/plugin-name: e2e-"${i}"/" > ./tmpPlugins/bucket${bucket}/plugin-$i.yaml
    cat prefixFile > tmpfile

    # Plugin for e2e starts with podspec and empty containers; we are redefining that in the prefix so remove those 2 lines.
    cat ./tmpPlugins/bucket${bucket}/plugin-$i.yaml | sed '1,2d' >> tmpfile && mv tmpfile ./tmpPlugins/bucket${bucket}/plugin-$i.yaml
    let i=i+1
done<focusList

rm focusList
rm tmptestlist

# Now we can run them in chunks that wont block long even with multiple failures and won't demand too many resources at once
for (( c=0; c<stop_at_bucket; c++ ))
do
  rm -rf ./tmpdir
  sonobuoy run -p ./tmpPlugins/bucket${c} --wait -n sonobuoy-iterative-bucket-${c}
  sonobuoy retrieve -x tmpdir -n sonobuoy-iterative-bucket-${c}

  # Output status to help with debug
  sonobuoy status --json -n sonobuoy-iterative-bucket-${c} | jq

  mkdir ${SONOBUOY_RESULTS_DIR}/bucket${c}logs
  for f in ./tmpdir/podlogs/*/*/logs/e2e.txt
  do
    base=$(basename "$f")
    mv "${f}" ${SONOBUOY_RESULTS_DIR}/bucket${c}logs/"${c}-${base}"
  done
  for f in ./tmpdir/podlogs/*/*/logs/sonobuoy-worker.txt
  do
      base=$(basename "$f")
      mv "${f}" ${SONOBUOY_RESULTS_DIR}/bucket${c}logs/"${c}-${base}"
  done
  for f in ./tmpdir/plugins/*/results/global/junit*
  do
      base=$(basename "$f")
      mv "${f}" ${SONOBUOY_RESULTS_DIR}/"${c}-${base}"
  done

  sonobuoy delete -n sonobuoy-iterative-bucket-${c}
done

mv ./tmpPlugins ${SONOBUOY_RESULTS_DIR}/tmpPlugins