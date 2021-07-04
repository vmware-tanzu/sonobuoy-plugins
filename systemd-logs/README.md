# Sonobuoy systemd-logs plugin

This is a simple standalone container that gathers log information from systemd, by chrooting into the node's filesystem and running `journalctl`.

This container is used by [Sonobuoy](https://github.com/vmware-tanzu/sonobuoy) for gathering host logs in a Kubernetes cluster.

You typically do not need to target this plugin manually since it is currently included in the sonobuoy CLI and runs automatically when you run:

```
sonobuoy run
```
