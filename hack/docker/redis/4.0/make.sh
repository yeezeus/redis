#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=redis
TAG=4.0
ALT_TAG=4
PATCH=4.0.6

docker pull $IMG:$PATCH-alpine

docker tag $IMG:$PATCH-alpine "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"

docker tag $IMG:$PATCH-alpine "$DOCKER_REGISTRY/$IMG:$ALT_TAG"
docker push "$DOCKER_REGISTRY/$IMG:$ALT_TAG"
