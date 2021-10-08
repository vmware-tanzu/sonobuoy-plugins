# Custom End-To-End (E2E) Tests

This plugin is meant as a skeleton for you to grab and run with to implement your
own custom tests in Kubernetes.

The benefits of using this plugin instead of starting from scratch:
 - Automatically comes with the [e2e-test-framework](https://github.com/kubernetes-sigs/e2e-framework) imported/configured
 - Includes basic examples so you don't have to look up basic boilerplate
 - Automatically comes with a Dockerfile and plugin.yaml so there is less overhead to getting started
 - Will get support as the e2e-test-framework and Sonobuoy evolve to get the best features supported by default

## How to use this plugin

- Clone this repo
- Modify the build script to specify your registry/image/tag
- Write tests (using main_test.go as a jumping off point)
- Run ./build.sh to build the image and push it to your registry
- `sonobuoy run -p plugin.yaml` to run your own plugin

## Roadmap:
 - Implement progress updates by default using the e2e-test-framework hooks
 - Within Sonobuoy, support `go test --json` output so that the results are intelligently parsed in order to get full Sonobuoy integration