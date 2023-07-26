package acceptancetests

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"regexp"
	"testing"
)

func TestAccNetworkingIpInterfaceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// non-existant SVM return code 2621462. Must happen before create/read
			{
				Config:      testAccNetworkingIPInterfaceResourceConfig("non-existant", "10.10.10.10", "ontap_cluster_1-01"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// non-existant homeport
			{
				Config:      testAccNetworkingIPInterfaceResourceConfig("carchi-test", "10.10.10.10", "non-existant_home_node"),
				ExpectError: regexp.MustCompile("393271"),
			},
			// Create and Read
			{
				Config: testAccNetworkingIPInterfaceResourceConfig("carchi-test", "10.10.10.10", "ontap_cluster_1-01"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "name", "test-interface"),
					resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "svm_name", "carchi-test"),
				),
			},
			// Update and Read (when update is implemented this is what it would look like)
			//{
			//	Config: testAccNetworkingIPInterfaceResourceConfig("carchi-test", "10.10.10.20"),
			//	Check: resource.ComposeTestCheckFunc(
			//		resource.TestCheckResourceAttr("netapp-ontap_networking_ip_interface_resource.example", "name", "test-interface", "ontap_cluster_1-01"),
			//	),
			//},
		},
	})
}

func testAccNetworkingIPInterfaceResourceConfig(svmName, address, homeNode string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
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
}
`, host, admin, password, svmName, address, homeNode)
}
