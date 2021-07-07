# e2e Plugin (Kubernetes End-To-End Tests)

### Introduction

The e2e plugin is the primary use case that [sonobuoy][sonobuoy] is known for. It runs the Kubernetes end-to-end tests on your targeted cluster.

Typically, there is no reason to target this YAML specifically as the sonobuoy CLI houses and generates it for you so that you can run this plugin simply by running:

```
sonobuoy run
```

And you can generate this plugin (with minor modifications) by running:

```
sonobuoy gen plugin e2e
```

### Windows Tests

Running e2e tests on Windows requires numerous modifications to this plugin and so it has its own so that running it is more simple.

See the [Windows plugin][windows] for more details and our [blog][windowsBlog] for a more thorough discussion.

### The Tests

The tests themselves are written by "upstream" Kubernetes 
and live in the [Kubernetes repo][kubernetes].

### Other Resources

See [sonobuoy.io][sonobuoySite] for additional documentation, blogs, and resources.

[sonobuoy]: https://github.com/vmware-tanzu/sonobuoy
[sonobuoySite]: https://sonobuoy.io/
[kubernetes]: https://github.com/kubernetes/kubernetes
[windows]: windows-e2e
[windowsBlog]: https://sonobuoy.io/windows-e2e-tests-with-sonobuoy/
