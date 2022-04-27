# Debug Workloads

An implementation of the [A visual guide on troubleshooting Kubernetes deployments](https://learnk8s.io/troubleshooting-deployments)

Runs a series of checks via Sonolark (a Sonobuoy plugin) to debug workloads in your cluster.

This is just a POC and developing it had a few goals:
 - act as a functional example you can build from
 - provided a means of finding gaps in the current functionality of Sonolark

Currently, this is incomplete (only about 40% automated) but this will hopefully be expanded as we fill in more of the gaps of Sonolark's capabilities and have the time to put here.

## How to use

If you have Sonolark locally, run via:

```
./hack/build.sh && ./hack/run.sh
```

To run as a Sonobuoy plugin:

```
./hack/build.sh && sonobuoy run -p plugin.yaml
```
