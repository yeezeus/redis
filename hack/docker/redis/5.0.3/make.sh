#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=redis
DB_VERSION=5.0.3
TAG="$DB_VERSION"

docker pull $IMG:$DB_VERSION-alpine

docker tag $IMG:$DB_VERSION-alpine "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"
