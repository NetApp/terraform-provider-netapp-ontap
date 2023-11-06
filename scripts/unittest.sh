#!/usr/bin/env bash

org_dir=$(pwd)

mkdir $org_dir/test
go test -buildvcs=false `go list ./... | grep -v acceptancetests` || echo "Build finished in error due to failed tests"
