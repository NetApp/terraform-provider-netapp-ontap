#!/usr/bin/env bash

org_dir=$(pwd)
export TF_ACC=1
#export TF_ACC_NETAPP_HOST="<Host1>"
#export TF_ACC_NETAPP_HOST2="<host2>>"
#export TF_ACC_NETAPP_USER="admin"
#export TF_ACC_NETAPP_PASS="<password>"
#export TF_ACC_NETAPP_LICENSE="<licensekey>"

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
