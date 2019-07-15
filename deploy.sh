#!/bin/bash
# Exit on any error
set -e
for f in manifests/*.yml
do
    envsubst < $f > "generated-$(basename $f)"
done
gcloud docker -- push asia.gcr.io/${PROJECT_NAME}/sample:$CIRCLE_SHA1
kubectl apply -f generated-deployment.yml --record
kubectl apply -f generated-service.yml