#!/bin/bash
# Exit on any error
set -e
for f in manifests/*.yaml
do
    envsubst < $f > "generated-$(basename $f)"
done
gcloud auth configure-docker
gcloud docker --verbosity=error -- push gcr.io/${PROJECT_NAME}/trade-derby:$CIRCLE_SHA1
kubectl apply -f generated-tradederby-deployment.yaml --record
kubectl apply -f generated-tradederby-service.yaml
# kubectl apply -f manifests/tradederby-deployment.yaml --record
# kubectl apply -f manifests/tradederby-service.yaml