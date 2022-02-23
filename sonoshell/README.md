# Sonoshell

This utility provides a mechanism for sonobuoy to run and test a suite of arbitrary commands, reporting the results of each of these back. It is intended to be consumed in a plugin, imported in the Dockerfile as the base image and provided a yaml file defining the suite of commands to run. An example, `test.yaml`, is provided as a guide. By default, sonoshell will run against it, demonstrating the pass and fail conditions, which are based on the exit code of the command, program or script defined for the test. An exit code of 0 indicates that a test has passed, any other exit code is treated as a failure.

Also included is an example plugin definition, `sonoshell.yaml`. This demonstrates a fully functional sonobuoy plugin which reports its results with sonoshell. It is configured to run as a `DaemonSet`, running against and reporting on every node in the cluster, but running it as a `Job` is also possible if the test suite only needs to run on one node.

The example plugin can be tested as follows (assuming running from an environment with a well-configured `sonobuoy` running against an active cluster):
```shell
sonobuoy gen --plugin sonoshell.yaml > sonoshell-run.yaml
# Any additional changes to the generated yaml per your environments needs go here
sonobuoy run -f sonoshell-run.yaml --wait

sonobuoy retrieve
```

To use sonoshell in another plugin, you can use it as a base image like so:
```dockerfile
FROM laevos/sonoshell:latest

COPY ./test-suite.yaml /test-suite.yaml

# Additional environment setup as necessary

CMD ["/sonoshell", "/test-suite.yaml"]
```
