# Requirements Check

>## _NOTE_ 
>This plugin is a work in progress. Please file an issue or join us in the #sonobuoy slack
channel to discuss or request features.
> 
>See the roadmap at the bottom of the readme for more details about the direction of this plugin.

The purpose of this plugin is to be able to check that a cluster meetst the
requirements of your apps/platforms/etc.

See the example plugin.yaml for how the input.json is attached to the plugin definition.

## How to use

1. Take the example plugin.yaml and modify the input.json data to document your own requirements.
2. Run `sonobuoy run -p yourplugin.yaml`

You typically wont need to modify the other details of the plugin since it uses the same
core logic/image regardless of your requirements input file.

## Roadmap

So far this is just a proof of concept of what this could grow into.

Our goals for growing this plugin:
- expanding the number of things we can check for
- expanding the output options so users can get the most out of the feedback
- increasing speed and reducing API calls by [probably] utilzing a `kubectl cluster-info dump` call once which we then reference repeatedly
- finding more use cases for checks that are not simply API data/state based. E.g. checking logs or more complex business logic
- improving documentation so it is easier to generate the input.json data