#!/bin/bash
docker build . -t schnake/postprocessor:v0
#docker push schnake/postprocessor:v0
kind load docker-image schnake/postprocessor:v0