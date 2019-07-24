#!/bin/bash
# Exit on any error
set -e

cd `dirname $0`

DEPLOYMENT_NAME="$1"
echo "deployment:" $DEPLOYMENT_NAME

/usr/local/bin/kubectl patch deployment $DEPLOYMENT_NAME -p \
  "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"reloaded-at\":\"`date +'%Y%m%d%H%M%S'`\"}}}}}"
