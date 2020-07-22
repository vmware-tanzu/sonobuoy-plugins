# CIS Benchmarks

This plugin utilizes the [kube-bench][kubebench] implementation of the [CIS security benchmarks][cis]. It is technically two plugins; one to run the checks on the master nodes and another to run the checks on the worker nodes.

## Usage

To run this plugin with the default options, run the following command:

```
sonobuoy run --plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/master/cis-benchmarks/kube-bench-plugin.yaml --plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/master/cis-benchmarks/kube-bench-master-plugin.yaml
```

The version of the CIS benchmark you should run depends on the version of your Kubernetes cluster.
The Kubernetes version can be set explicitly when running kube-bench or kube-bench can auto-detect the version if you mount `kubectl` or `kubelet` into the container running it.

By default, this plugin explicitly sets the version and assumes you are running Kubernetes 1.17.
You can modify this and set the version when running by setting the `KUBERNETES_VERSION` environment variable for each of the two plugins by adding the following flags:
`--plugin-env kube-bench-master.KUBERNETES_VERSION=<version> --plugin-env kube-bench-node.KUBERNETES_VERSION=<version>`.
You can also download and modify the plugin YAML definitions directly.

If you wish to the use the Kubernetes version auto-detection feature, you must uncomment the [volume definition](./kube-bench-master-plugin.yaml#L24-L26) and [volume mount](./kube-bench-master-plugin.yaml#L73-L74) in both plugin definitions.
This will result in the `/usr/bin` directory on the nodes being mounted into the container.

**NOTE**: Some users have experienced issues when using this approach.
This has been fixed in kube-bench however hasn't been released yet.
For simplicity, we recommend specifying the version explicitly using the instructions above.

## Plugin options

The plugin can be configured by setting a number of options using environment variables.
For all of the environment described below, they can be set by modifying the value in the plugin YAML, or can be set using Sonobuoy's `--plugin-env` flag as follows:
* `--plugin-env kube-bench-master.ENV_VAR=<value>` for the kube-bench-master plugin, or
* `--plugin-env kube-bench-node.ENV_VAR=<value>` for the kube-bench-node plugin

### Environment variable options
* `KUBERNETES_VERSION`
  This can be set to specify the Kubernetes version of your cluster. This is used to determine which version of the CIS benchmark kube-bench will run.
* `DISTRIBUTION`
  This can be set if the default configuration for kube-bench is not compatible with the Kubernetes distribution you are using.
  By setting this value, a distribution specific configuration will be used when running kube-bench.
  The supported distributions are [Enterprise PKS (`entpks`)](./entpks), [Google Kubernetes Engine (`gke`)](./gke) and [Elastic Kubernetes Service (`eks)](./eks).

The following environment variables should only be modified if your cluster is Kubernetes v1.15+ and as such will be running version 1.5 of the CIS benchmark.
The default settings for these environment variables are compatible with all versions of the benchmark.
CIS 1.5 introduces a number of new targets to run checks for rather than just the master and node targets.
Each of targets can be enabled or disabled by setting the value for the appropriate variable to "true" or "false".

* `TARGET_MASTER`
  Setting this to "true" enables the checks for master nodes. For all versions of the CIS benchmark, this is Section 1.
  This is enabled by default in the kube-bench-master plugin.
* `TARGET_NODE`
  Setting this to "true" enables the checks for worker nodes. For CIS 1.5, this is Section 4. For all other versions of the CIS benchmark, this is Section 2.
  This is enabled by default in the kube-bench-node plugin.
* `TARGET_ETCD`
  Setting this to "true" enables the checks for etcd configuration. For CIS 1.5, this is Section 2. This target cannot be enabled for earlier versions of the benchmark.
  This is disabled by default in both plugins.
* `TARGET_CONTROLPLANE`
  Setting this to "true" enables the checks for control plane configuration. For CIS 1.5, this is Section 3. This target cannot be enabled for earlier versions of the benchmark.
  This is disabled by default in both plugins.
* `TARGET_POLICIES`
  Setting this to "true" enables the checks for Kubernetes policies. For CIS 1.5, this is Section 2. This target cannot be enabled for earlier versions of the benchmark.
  This is disabled by default in both plugins.

The following environment variables are distribution specific.
They should not be enabled unless you are running against a distribution where they are valid and compatible.
Enabling these when running against an unsupported distribution will result in the plugin failing to run correctly.

* `TARGET_MANAGED_SERVICES`
  Setting this to "true" enables the checks for managed service components.
  This target is only available when running the [CIS GKE benchmark](./gke).
  This is enabled by default for the GKE version of the plugin.

## Distribution specific support

Some Kubernetes distributions require custom configurations to be provided in order to run kube-bench correctly.
To perform the checks listed in the CIS benchmark, it is necessary for kube-bench to know the locations on disk of various Kubernetes configuration files.
If the paths for a distribution's configuration files are not included in the list of [default locations](https://github.com/aquasecurity/kube-bench/blob/master/cfg/config.yaml), these must be provided via the kube-bench configuration.

This plugin will include custom configuration for different distributions.
Currently, only Enterprise PKS is supported as a custom distribution.
If a custom distribution is not specified as described above, the default configuration for kube-bench will be used.

## Assumptions

To run both plugins (with the command above) the following assumptions are made:

 - One or more master node (with the label `node-role.kubernetes.io/master`)
 - One or more worker node (without the master node label)
 - Sonobuoy 0.16.4 (relies on support for node affinity and the command above expects `--plugin` to take a URL)

If you just want to run one or the other checks, specify only one of the plugins rather than both.

## Customization

Although you can run the plugins by specifying the URL for the YAML in this repository, you can also download the YAML and modify it if you need a custom mount or would like to specify other options to the kube-bench application.

[kubebench]: https://github.com/aquasecurity/kube-bench
[cis]: https://www.cisecurity.org/benchmark/kubernetes
