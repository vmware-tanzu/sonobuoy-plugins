#!/bin/bash
docker build . -t sonobuoy/requirementscheck:v0.0.1
docker push sonobuoy/requirementscheck:v0.0.1