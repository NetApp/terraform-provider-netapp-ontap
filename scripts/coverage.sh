#!/usr/bin/env bash

org_dir=$(pwd)

rm -rf $org_dir/test
mkdir $org_dir/test
go test -coverprofile $org_dir/test/cover.out `go list ./... | grep -v acceptancetests`
if [ $? -eq 0 ]
then
    go tool cover -html=$org_dir/test/cover.out
else
    echo "Build finished in error due to failed tests"
    exit 1
fi
