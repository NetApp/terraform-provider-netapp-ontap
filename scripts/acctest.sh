#!/usr/bin/env bash

org_dir=$(pwd)
export TF_ACC=1
export TF_ACC_NETAPP_HOST="10.193.180.108"
export TF_ACC_NETAPP_HOST2="10.193.176.186"
export TF_ACC_NETAPP_HOST3="10.193.176.186"
export TF_ACC_NETAPP_USER="admin"
export TF_ACC_NETAPP_PASS="netapp1!"
export TF_ACC_NETAPP_LICENSE="SMEXXDBBVAAAAAAAAAAAAAAAAAAA"

rm -rf $org_dir/test
mkdir $org_dir/test
go test -cover -coverprofile $org_dir/test/cover.out `go list ./... | grep -e provider`
if [ $? -eq 0 ]
then
    go tool cover -html=$org_dir/test/cover.out
else
    echo "Build finished in error due to failed tests"
    exit 1
fi
