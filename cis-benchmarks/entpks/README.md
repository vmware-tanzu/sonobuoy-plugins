# CIS Benchmark for Enterprise PKS

This directory contains an adapted version of the CIS Benchmark plugin to be used with Enterprise PKS.
This includes both the plugin definition and a customized configuration to be used with kube-bench that details where the configuration files for each of the components can be found.

Enterprise PKS does not support running workloads on master nodes within a cluster.
Given that Sonobuoy runs its plugins as Pods or DaemonSets, it is not possible to run the "kube-bench-master" plugin.
Only the plugin for running on worker nodes is provided here.
However, if you wish to run the CIS benchmark on your master nodes, you can still make use of the custom configuration that is provided within this directory.

To do this, you will need to be able to SSH into your master nodes and download the [latest release of kube-bench](https://github.com/aquasecurity/kube-bench/releases).
You will also need to either clone this repository, or download the [config.yaml](./config.yaml) file.
Once you have installed the tool, follow the [instructions for running kube-bench](https://github.com/aquasecurity/kube-bench#running-kube-bench), and provide the custom configuration file using the `--config <path-to-custom-config>` flag.
