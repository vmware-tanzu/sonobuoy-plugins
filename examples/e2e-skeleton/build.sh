#!/usr/bin/env bash

REGISTRY=schnake
IMG=custome2e
TAG=v0.2.0

docker build . -t $REGISTRY/$IMG:$TAG
docker push $REGISTRY/$IMG:$TAG