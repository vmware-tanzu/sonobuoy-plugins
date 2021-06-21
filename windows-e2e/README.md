# Windows E2E Tests

There are three plugins in this repo that correspond to different Windows/Kubernetes versions. See [windows-testing][windowsTestingRegistries] documentation for which to use.

The reason for the different versions is that different versions of Kubernetes have slightly different testing requirements.

To use, just run:

```
sonobuoy run -p <url>
```

or download it and use the filename instead of URL.

For more details about how the plugin differs from the typical e2e plugin, check our [blog][blog] and get more details from [windows-testing][windowsTesting].

[windowsTesting]: https://github.com/kubernetes-sigs/windows-testing
[blog]: https://sonobuoy.io/blog/
[windowsTestingRegistries]: https://github.com/kubernetes-sigs/windows-testing/tree/master/images
