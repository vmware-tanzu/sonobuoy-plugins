# Sonolark

> Note: This plugin is currently a Work In Progress (WIP) and currently under development.

The goal of this plugin is to create offer a plugin which:
 - is highly (and easily) configurable
 - does not require the user to build/manage images
 - makes working with Kubernetes (k8s) resources easy
 - integrates easily with Sonobuoy

We accomplish this by providing a library of helper functions baked into a prebuilt image which the user calls via a Starlark script they provide.

## How to make your own

Two main appraoches:
 - Start with plugin.yaml and simply edit the script.star which is provided via a configmap
 - Write your own script.star and use `sonobuoy gen plugin` along with the `--configmap` flag to target your script

```bash
sonobuoy gen plugin --name=sonolark --image=vmware-tanzu/sonolark:v0.0.1 --configmap=./script.star --format=manual -c "./sonolark" > plugin.yaml
```

The benefit of the latter approach is that your script.star file will have normal indentation instead of the extra padding caused by being placed into the yaml file.