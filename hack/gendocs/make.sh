#!/usr/bin/env bash

pushd $GOPATH/src/kubedb.dev/redis/hack/gendocs
go run main.go
popd
