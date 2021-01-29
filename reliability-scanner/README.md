# Reliability Scanner

The Reliability Scanner is a customizable Sonobuoy Plugin that captures good practices for operating workloads reliably atop Kubernetes.

The plugin is aimed at providing a suggestive set of checks, that can be conditionally included and configured to an end-users liking. This will help to identify workloads or cluster configuration that do not make best use of the relability features offered by a particular platform, or it's extended tooling.

The resulting report, or status of a completed scan can assist an end-user to identify misaligned configuration or highlight areas that may have a negative impact on organizational reliability targets.

## Getting Started

### Dependencies
[Sonobuoy](https://github.com/vmware-tanzu/sonobuoy)
[YTT](https://github.com/vmware-tanzu/carvel-ytt)
[GNU Make](https://www.gnu.org/software/make/)

Currently, the plugin is distributed as a docker container. The Reliability Scanner requires YTT as a dependancy for configuration templating.

### Quickstart

The default Reliability Scanner configuration file can be found at `./plugin/reliability-scanner-custom-values.lib.yml`. This configuration file may be used to include and configure checks for a Scan.

This file may be customized to include and configure checks supported by the Reliability Scanner. See [Customizing Checks](#customizing-checks).

A Sonobuoy run with the Reliability Scanner enabled may be initialized using.

```
make run
```

Results may be viewed using the standard Sonobuoy retrieve command. (see `./Makefile`)

```
make results
```

## Customizing Checks

Checks currently available through the Reliability Scanner are as follows.

| Version  | Kind    | Check       | Parameter                 | Description                                                                | Type                                            | Default        |
|----------|---------|-------------|---------------------------|----------------------------------------------------------------------------|-------------------------------------------------|----------------|
| v1alpha1 | pod     | qos         | minimum_desired_qos_class | Defines the minimum desired QOS class for Pods running within the cluster. | String, ["BestEffort"/"Burstable"/"Guaranteed"] | "BestEffort"   |
|          |         |             | include_detail            | Include detail of the current configured QOS class per Pod.                | Boolean, [true/false]                           | true           |
| v1alpha1 | pod     | probes      | -                         | Checks if Pod Liveness/Readiness Probes are defined.                       | -                                               | -              |
| v1alpha1 | service | annotations | key                       | Checks for a specific annotation key on a service kind.                    | String                                          | "incident-url" |
|          |         |             | include_annotations       | Include currently configured annotations.                                  | Boolean, [true/false]                           | true           |                                                         

Each check may be conditionally included and customized to suit the requirements of the target cluster. The default set of checks are defined in `./plugin/reliability-scanner-custom-values.lib.yml`.
