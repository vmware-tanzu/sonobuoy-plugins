#!/usr/bin/env bash

# Various tests cases. Messy but pragmatic during initial iteration.
# Should make actual tests.
# kubectl apply -f ./tests/test_happypath.yaml
# kubectl apply -f ./tests/test_ingressissuename.yaml
# kubectl apply -f ./tests/test_ingressissueport.yaml
# kubectl apply -f ./tests/test_serviceissuename.yaml
# kubectl apply -f ./tests/test_serviceissueport.yaml
# Cant do the below because k8s will reject the update
# kubectl apply -f ./tests/test_deploymentlabelissue.yaml
# kubectl apply -f ./tests/test_clusterfull.yaml
# kubectl apply -f ./tests/test_resourcequotaissue.yaml
# kubectl apply -f ./tests/test_pvcDNEissue.yaml
# kubectl apply -f ./tests/test_pvcNotReadyissue.yaml

# sonobuoy run -p plugin.yaml --wait
SONOLARK_DEPLOYMENT=default/my-deployment \
 SONOLARK_INGRESS=default/my-ingress \
 SONOLARK_SERVICE=default/my-service \
 sonolark -f script.star
