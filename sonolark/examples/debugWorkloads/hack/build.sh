#!/usr/bin/env bash

# Uses a tag that hasn't been published yet.
sonobuoy gen plugin \
 -n debugworkloads \
 -i sonobuoy/sonolark:v0.0.3 \
 -e SONOLARK_DEPLOYMENT=default/my-deployment \
 -e SONOLARK_INGRESS=default/my-ingress \
 -e SONOLARK_SERVICE=default/my-service \
 --configmap script.star \
 --format manual | \
 yq 'del(.spec.command)'| \
 yq 'del(.spec.resources)'| \
 yq 'del(.spec.volumeMounts)' > plugin.yaml
 