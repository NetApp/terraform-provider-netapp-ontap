
# set -x
if [ ! -d internal/provider/ ]; then
    echo "This script creates the source file skeleton for a resource."
    echo "This script needs to be run from the top of the tree"
    echo "current location: "
    pwd
    exit 1
fi

function link_if_exist() {
    source=$1
    if [ -e $source ]; then
        ln -s $source .
    else
        echo "Error: $source does not exist!"
    fi
}

echo "Resource name: eg ip_interface"
read tag_prefix
echo "Resource GO prefix: eg IPInterface"
read go_prefix

# resource file
provider_file=internal/provider/${tag_prefix}_resource.go
bad_provider_file=internal/provider/${tag_prefix}_resource.go-e
if [ -e  $provider_file ]; then
    echo "resource file $provider_file already exists"
else
    echo "creating $provider_file"
    cp internal/provider/tag_prefix_resource.go $provider_file
    sed -i'' -e s/tag_prefix/${tag_prefix}/g $provider_file
    sed -i'' -e s/GoPrefix/${go_prefix}/g $provider_file
    rm -rf $bad_provider_file
fi

# The same file is used for both data sources and resource
interface_file=internal/interfaces/${tag_prefix}.go
bad_provider_file=internal/interfaces/${tag_prefix}.go-e
if [ -e $interface_file ]; then
    echo "interfaces file $interface_file already exists"
else
    echo "creating $interface_file"
    cp internal/interfaces/tag_prefix.go $interface_file
fi
    sed -i'' -e s/tag_prefix/${tag_prefix}/g $interface_file
    sed -i'' -e s/GoPrefix/${go_prefix}/g $interface_file
    rm -rf $bad_provider_file

example_dir=examples/resources/netapp-ontap_${tag_prefix}
if [ -d $example_dir ]; then
    echo "example dir $example_dir already exists"
else
    provider_dir=../../provider
    echo "creating $example_dir"
    mkdir $example_dir
    cd $example_dir
    link_if_exist ../../provider/provider.tf
    link_if_exist ../../provider/variables.tf
    link_if_exist ../../provider/terraform.tfvars
    cat > resource.tf << EOF
reource "netapp-ontap_${tag_prefix}_resource" "${tag_prefix}" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "testme"
}
EOF
fi

echo "TODO: update internal/provider/provider.go to register new data source New${go_prefix}Resource"
echo "TODO: add Unit Tests"
