#!/usr/bin/env bash

set -eoux pipefail

REPO_NAME=redis
OPERATOR_NAME=rd-operator

# get concourse-common
pushd $REPO_NAME
git status
git subtree pull --prefix hack/concourse/common https://github.com/kubedb/concourse-common.git master --squash -m 'concourse'
popd

source $REPO_NAME/hack/concourse/common/init.sh

pushd "$GOPATH"/src/github.com/kubedb/$REPO_NAME

# build and push docker-image
./hack/builddeps.sh
export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci

./hack/docker/rd-operator/make.sh build
./hack/docker/rd-operator/make.sh push
popd

pushd $GOPATH/src/github.com/kubedb/$REPO_NAME

# run tests
source ./hack/deploy/make.sh --docker-registry=kubedbci
./hack/make.py test e2e --v=1 --storageclass=$StorageClass --selfhosted-operator=true --ginkgo.flakeAttempts=2
