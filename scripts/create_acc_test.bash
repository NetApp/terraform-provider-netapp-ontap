#!/bin/bash

if [ ! -d internal/provider/ ]; then
    echo "This script creates the test file skeleton for a resource."
    echo "This script needs to be run from the top of the tree"
    echo "current location: "
    pwd
    exit 1
fi

echo "Resource name for test: eg snapmirror"
read new_test_file_name
echo "Resource GO prefix: eg SnapmirrorResource"
read go_prefix_name

# Define the existing resource name
existing_resource_name="snapmirror_resource"

# Get the new resource name from the command-line argument
new_resource_name=${new_test_file_name}_resource

# Get the go prefix from the command-line argument
go_prefix=${go_prefix_name}

# Define the path to the test directory
test_directory="internal/provider"

# Define the path to the new test file
new_test_file="${test_directory}/${new_resource_name}_test.go"

# Check if the new test file already exists
if [[ -f ${new_test_file} ]]; then
    echo "Error: File ${new_test_file} already exists."
    exit 1
fi

bad_test_file=internal/provider/${new_resource_name}_test.go-e
echo "creating $new_test_file"

# Copy the existing test file to create a new test file
cp "${test_directory}/${existing_resource_name}_test.go" "${new_test_file}"

# Replace all occurrences of the existing resource name with the new resource name in the new test file
sed -i -e "s/${existing_resource_name}/${new_resource_name}/g" "${new_test_file}"

sed -i -e "113,114d;110,111d;107,108d;78,79d;75,76d;17,21d;6d;" "${new_test_file}"
sed -i -e "s/SnapmirrorResource/${go_prefix}/g" "${new_test_file}"
sed -i -e "s/snapmirror_dest_svm:testme/name/g" "${new_test_file}"
sed -i -e "s/snapmirror_source_svm:snap3/acc_test/g" "${new_test_file}"
sed -i -e "s/snapmirror_source_svm:testme/svm_name/g" "${new_test_file}"
sed -i -e "s/snapmirror_dest_svm:snap_dest/name/g" "${new_test_file}"
sed -i -e "s/snapmirror_source_svm:snap/svm_name/g" "${new_test_file}"
sed -i -e "s/destination_endpoint.path/name/g" "${new_test_file}"
sed -i -e "s/policy.name/option_name/g" "${new_test_file}"
sed -i -e "s/MirrorAndVault/option/g" "${new_test_file}"
sed -i -e "s/sourceEndpoint/name/g" "${new_test_file}"
sed -i -e "s/destinationEndpoint/svmName/g" "${new_test_file}"
sed -i -e "s/source_endpoint = {/name = \"%s\"/g" "${new_test_file}"
sed -i -e "s/destination_endpoint = {/svm_name = \"%s\"/g" "${new_test_file}"
sed -i -e "s/policy = {/option_name = \"%s\"/g" "${new_test_file}"
sed -i -e "s/policy/option/g" "${new_test_file}"
sed -i -e "s/snapmirror/${new_test_file_name}/g" "${new_test_file}"

rm -rf $bad_test_file

echo "Done."