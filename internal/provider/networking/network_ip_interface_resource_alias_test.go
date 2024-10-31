package networking_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetworkIpInterfaceResourceAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// non-existant SVM return code 2621462. Must happen before create/read
			{
				Config:      testAccNetworkIPInterfaceResourceConfigAlias("non-existant", "10.10.10.10", "ontap_cluster_1-01", "default-data-files"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// non-existant home node
			{
				Config:      testAccNetworkIPInterfaceResourceConfigAlias("terraform", "10.10.10.10", "non-existant_home_node", "default-data-files"),
				ExpectError: regexp.MustCompile("53281680"),
			},
			// Create and Read
			// {
			// 	Config: testAccNetworkIPInterfaceResourceConfigAlias("terraform", "10.10.10.10", "ontap_cluster_1-01", "default-data-files"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "name", "test-interface"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "svm_name", "terraform"),
			//    resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "service_policy", "default-data-files"),
			// 	),
			// },
			// // Update and Read
			// {
			// 	Config: testAccNetworkIPInterfaceResourceConfigAlias("terraform", "10.10.10.20", "ontap_cluster_1-01", "default-data-iscsi"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "name", "test-interface"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "ip.address", "10.10.10.20"),
			//    resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "service_policy", "default-data-iscsi"),
			// 	),
			// },
			// Test importing a resource
			// {
			// 	ResourceName:  "netapp-ontap_networking_ip_interface_resource.example",
			// 	ImportState:   true,
			// 	ImportStateId: fmt.Sprintf("%s,%s,%s", "acc_test", "terraform", "cluster4"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "name", "acc_test"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "ip.address", "10.193.180.56"),
			// 	),
			// },
		},
	})
}

func testAccNetworkIPInterfaceResourceConfigAlias(svmName, address, homeNode, servicePolicy string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST5")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST5, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_networking_ip_interface_resource" "example" {
	cx_profile_name = "cluster4"
	name = "test-interface"
	svm_name = "%s"
  	ip = {
    	address = "%s"
    	netmask = 18
    }
  	location = {
    	home_port = "e0d"
    	home_node = "%s"
  	}
	service_policy = "%s"
}
`, host, admin, password, svmName, address, homeNode, servicePolicy)
}
