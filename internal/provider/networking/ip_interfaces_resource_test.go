package networking_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetworkIpInterfacesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// non-existant SVM return code 2621462. Must happen before create/read
			{
				Config:      testAccNetworkIPInterfaceResourceConfig("non-existant", "10.10.10.10", "ontap_cluster_1-01"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// non-existant home node
			{
				Config:      testAccNetworkIPInterfaceResourceConfig("svm0", "10.10.10.10", "non-existant_home_node"),
				ExpectError: regexp.MustCompile("393271"),
			},
			// Create and Read
			{
				Config: testAccNetworkIPInterfaceResourceConfig("svm0", "10.10.10.10", "ontap_cluster_1-01"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interfaces.example", "name", "test-interface"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interfaces.example", "svm_name", "svm0"),
				),
			},
			// Update and Read
			{
				Config: testAccNetworkIPInterfaceResourceConfig("svm0", "10.10.10.20", "ontap_cluster_1-01"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interfaces.example", "name", "test-interface"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interfaces.example", "ip.address", "10.10.10.20"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_network_ip_interfaces.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "test-interface", "svm0", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interfaces.example", "name", "test-interface"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interfaces.example", "ip.address", "10.10.10.20"),
				),
			},
		},
	})
}

func testAccNetworkIPInterfaceResourceConfig(svmName, address, homeNode string) string {
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

resource "netapp-ontap_network_ip_interfaces" "example" {
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
