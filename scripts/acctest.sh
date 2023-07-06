#!/usr/bin/env bash

org_dir=$(pwd)

mkdir $org_dir/test
TF_ACC=1 go test github.com/netapp/terraform-provider-netapp-ontap/internal/provider/acceptancetests -v || { echo "Build finished in error due to failed tests" && exit 1; }
