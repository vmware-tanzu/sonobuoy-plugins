#!/usr/bin/env bash

kubectl delete deployments --all
kubectl delete pods --all
kubectl delete service --all
kubectl delete ingress --all
kubectl delete resourcequota --all