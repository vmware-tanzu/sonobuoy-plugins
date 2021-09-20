#!/bin/bash
docker build . -t schnake/requirementscheck:v0
docker push schnake/requirementscheck:v0