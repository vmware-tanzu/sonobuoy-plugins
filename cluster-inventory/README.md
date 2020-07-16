# Cluster Inventory

This plugin allows you collect information about your cluster, such as operational details and details about workloads that are running on the cluster across all namespaces.

The plugin queries the cluster to collect the following information:
 * Cluster node details
 * Control Plane details
 * CNI details (version, plugins, etc.)
 * Cluster namespaces
 * Workloads running in each namespace:
    * Deployments
    * ReplicaSets
    * ReplicationControllers
    * StatefulSets
    * DaemonSets
    * CronJobs
    * Jobs
    * Pods

The workloads gathered by the plugin are presented in a tree structure showing the owner relationships.
For each object inspected by the plugin, its ownerReferences are resolved and the ownership tree is constructed accordingly.
This enables users of the plugin to explore the tree of workloads in their cluster.
For example, a user can explore the ReplicaSets created by a Deployment, and the Pods created by that ReplicaSet.


## Prerequisites
This plugin uses features that are only available in [Sonobuoy v0.18.1](https://github.com/vmware-tanzu/sonobuoy/releases/tag/v0.18.1).
It will still run with earlier versions of Sonobuoy however, we recommend using the latest available release to ensure the results are reported correctly to enable the full functionality when using the `sonobuoy results` command.

## Usage

To run this plugin, run the following command:

```
sonobuoy run --plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/master/cluster-inventory/cluster-inventory.yaml
``` 

You can check the status of the plugin, by running:

```
sonobuoy status
```

Once the plugin is complete, you can retrieve and view the results as follows:

```
results=$(sonobuoy retrieve)
sonobuoy results $results --mode dump
```

This will output the full report in the [Sonobuoy results format](https://sonobuoy.io/docs/results/).

## Report formats

The plugin is capable of producing two different report formats:
 * Sonobuoy report (`--sonobuoy-report`)
 * JSON report (`--json-report`)
 
### Sonobuoy report

This report presents information collected in the [Sonobuoy results format](https://sonobuoy.io/docs/results/).
This enables the plugin results to be processed and presented through Sonobuoy.

You can view a sample report [here](https://gist.github.com/zubron/242f128fd311e394853e3fdb339f7710).

The Sonobuoy results format is a recursive tree-like structure.
Each object in the report includes the following information:
 * Name
 * Status
 * Metadata
 * Details
 * Items (objects owned by this object)


### JSON report

This report simply marshalls all the information collected by the plugin into JSON.
Unlike the Sonobuoy results report, the data is presented as is, and in most cases will be the raw Kubernetes object.

You can view a sample report [here](https://gist.github.com/zubron/962f87e8d4e226b77ca08f456f866166).
