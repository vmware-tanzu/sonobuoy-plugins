# CIS Benchmark for Elastic Kubernetes Service

This directory contains an adapted version of the CIS Benchmark plugin to be used with Elastic Kubernetes Service (EKS).
Running the plugin on EKS does not require any additional configuration files, just the adapted plugin definition.

EKS does not provide access to master nodes within a cluster.
Due to this, it is not possible to run the "kube-bench-master" plugin.
Only the plugin for running on worker nodes is provided here which runs the `node`, `policies`, and `managedservices` targets.

Which version of the CIS benchmark to run depends on the version of your cluster.
For clusters with a version lower than v1.15, the standard version of the CIS benchmark for that version should be used.
For clusters where the version is v1.15 or later, the custom EKS benchmark should be used.
The Sonobuoy plugin provided here will determine which benchmark to use based on the Kubernetes version provided.
If the environment variable `KUBERNETES_VERSION` is not set, or is v1.15 or greater, the custom EKS benchmark will be run, otherwise, the benchmark matching the version will be run.
