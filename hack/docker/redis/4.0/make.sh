#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=redis
TAG=4.0
ALT_TAG=4

docker pull $IMG:$TAG-alpine

docker tag $IMG:$TAG-alpine "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"

docker tag $IMG:$TAG-alpine "$DOCKER_REGISTRY/$IMG:$TAG-alpine"
docker push "$DOCKER_REGISTRY/$IMG:$TAG-alpine"

docker tag $IMG:$TAG-alpine "$DOCKER_REGISTRY/$IMG:$ALT_TAG"
docker push "$DOCKER_REGISTRY/$IMG:$ALT_TAG"

docker tag $IMG:$TAG-alpine "$DOCKER_REGISTRY/$IMG:$ALT_TAG-alpine"
docker push "$DOCKER_REGISTRY/$IMG:$ALT_TAG-alpine"
