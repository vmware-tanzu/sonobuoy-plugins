# ytt Post-processor

>## _NOTE_ 
>This is a work in progress. Please file an issue or join us in the #sonobuoy slack
channel to discuss or request features.

The purpose of this image/package is to be used alongside your own Sonobuoy plugin to add common post-processing capabilities.
The core power of this plugin comes from the amazing transformation abilities of [ytt][ytt].
See their repo and documentation for details on how to utilize it.

See the examples directory for how to add this into your plugin.

## Why ytt?

Originally this idea wasn't bound to ytt at all.
We intended on creating a plugin that would accomoplish a host of post-processing steps.
As coding continued, we realized that ytt overlways could do all of the core things we were trying to accomplish and even more that we hadn't considered.

## Use cases

- Add context to test results (links to KB articles, useful debugging tips, etc)
- Remove tests which fail/warn users but which are not applicable to your system
- Remove the thousands of skipped tests from the e2e results so it is easier to read
- Add custom keywords to output to make it easier for your system to search for or process them

Let us know if you decide to use this for other use cases not mentioned here.

## How to use

1. Use the plugin.yaml file as an example of how to add this image as a sidecar container to your plugin.
   1. It needs to have a config-map of transforms. These files are fed to ytt and modify the plugin results
   2. The PodSpec should specify the new, postprocessing container
   3. The plugin should use the 'manual' results format since that's what the post-procressor creates 
2. Run `sonobuoy run -p yourplugin.yaml`

## How it works

1. First, the post-processor will wait for the 'done' file from the plugin, reporting that it is complete.
2. Second, the post-processor removes that 'done' file so that the 'sonobuoy-worker' container will not upload results.
3. Third, the post-processor transforms the plugin results (which may be in gojson or junit format) and transform them into the canonical Sonobuoy yaml format.
4. Finally, we invoke ytt with the files that you provided (via configmaps) to transform the data.

[ytt]: https://carvel.dev/ytt/