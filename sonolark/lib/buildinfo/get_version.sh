#!/bin/bash
git rev-parse --verify HEAD | tr -d '\n' > gitsha.txt

# This is THE source of truth for the version (binary and image)
echo -n v0.0.1 > version.txt