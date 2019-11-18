# CIS Benchmarks

This plugin utilizes the [kube-bench][kubebench] implementation of the [CIS security benchmarks][cis]. It is technically two plugins; one to run the checks on the master nodes and another to run the checks on the worker nodes.

## Usage

To run this plugin, run the following command:

```
sonobuoy run
--plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/cis-benchmarks/cis-benchmarks/kube-bench-plugin.yaml --plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/cis-benchmarks/cis-benchmarks/kube-bench-master-plugin.yaml 
```

## Assumptions

To run both plugins (with the command above) the following assumptions are made:

 - One or more master node (with the label `node-role.kubernetes.io/master`)
 - One or more worker node (without the master node label)
 - Using Kubernetes 1.13+
 - Sonobuoy 0.16.4 (relies on support for node affinity and the command above expects `--plugin` to take a URL)

If you just want to run one or the other checks, specify only one of the plugins rather than both.

## Customization

Although you can run the plugins by specifying the URL for the YAML in this repository, you can also download the YAML and modify it if you need a custom mount or would like to specify other options to the kube-bench application.

[kubebench]: https://github.com/aquasecurity/kube-bench
[cis]: https://www.cisecurity.org/benchmark/kubernetes