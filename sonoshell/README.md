# Sonoshell

This utility provides a mechanism for sonobuoy to run and test a suite of arbitrary commands, reporting the results of each of these back. A config-map containing the suite of tests to run should be included in the plugin definition. An example plugin definition, `sonoshell.yaml`, is included for reference. It demonstrates a fully functional sonobuoy plugin which reoprts its results with sonoshell. Each test will be invoked as a `bash` script. An exit code of 0 indicates that a test has passed, any other exit code is treated as a failure.

In the example plugin definition, sonoshell is configured to run as a `DaemonSet`, running against and reporting on every node in the cluster, but running it as a `Job` is also possible if the test suite only needs to run on one node.

The example plugin can be tested as follows (assuming running from an environment with a well-configured `sonobuoy` running against an active cluster):
```shell
sonobuoy gen --plugin sonoshell.yaml > sonoshell-run.yaml
# Any additional changes to the generated yaml per your environments needs go here
sonobuoy run -f sonoshell-run.yaml --wait

sonobuoy retrieve
```
