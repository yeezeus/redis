#!/usr/bin/env bash

set -eoux pipefail

ORG_NAME=kubedb
REPO_NAME=redis
OPERATOR_NAME=rd-operator
APP_LABEL=kubedb #required for `kubectl describe deploy -n kube-system -l app=$APP_LABEL`

export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci

# get concourse-common
pushd $REPO_NAME
git status # required, otherwise you'll get error `Working tree has modifications.  Cannot add.`. why?
git subtree pull --prefix hack/libbuild https://github.com/appscodelabs/libbuild.git master --squash -m 'concourse'
popd

source $REPO_NAME/hack/libbuild/concourse/init.sh

pushd "$GOPATH"/src/github.com/kubedb/$REPO_NAME

# build and push docker-image
./hack/builddeps.sh

./hack/docker/$OPERATOR_NAME/make.sh build
./hack/docker/$OPERATOR_NAME/make.sh push

# run tests
source ./hack/deploy/setup.sh --docker-registry=kubedbci
./hack/make.py test e2e --v=1 --storageclass=$StorageClass --selfhosted-operator=true --ginkgo.flakeAttempts=2
