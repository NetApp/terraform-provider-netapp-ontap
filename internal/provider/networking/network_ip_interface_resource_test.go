
package networking_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// example ID: aeef4e4f-a663-11ef-9ca8-00a0b8bc0407
const idRegexNetworkIPInterface string = "[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}"

func TestAccNetworkIpInterfaceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// non-existant SVM return code 2621462. Must happen before create/read
			{
				Config:      testAccNetworkIPInterfaceResourceConfigHomePortNode("non-existant", "10.10.10.10", "e0d", "ontap_cluster_1-01", "default-data-files", "Default"),
				ExpectError: regexp.MustCompile("Code:\"2621462\""),
			},
			// non-existant home node
			{
				Config:      testAccNetworkIPInterfaceResourceConfigHomePortNode("svm0", "10.10.10.10", "e0d", "non-existant_home_node", "default-data-files", "Default"),
				ExpectError: regexp.MustCompile("Code:\"53281680\""),
			},
			// non-existant broadcast domain
			{
				Config:      testAccNetworkIPInterfaceResourceConfigBroadcastDomain("svm0", "10.10.10.10", "non-existant_broadcast_domain", "default-data-files", "Default"),
				ExpectError: regexp.MustCompile("Code:\"2\""),
			},
			// empty location, no Home Node / Home Port / Broadcast Domain
			// Error 1967111: "Home node must be specified by at least one location.home_node, location.home_port, or location.broadcast_domain field."
			{
				Config:      testAccNetworkIPInterfaceResourceConfigHomePortNode("svm0", "10.10.10.10", "", "", "default-data-files", "Default"),
				ExpectError: regexp.MustCompile("Code:\"1967111\""),
			},

			// Create and Read (with Home Port & Node)
			{
				Config: testAccNetworkIPInterfaceResourceConfigHomePortNode("svm0", "10.10.10.10", "e0d", "ontap_cluster_1-01", "default-data-files", "Default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "name", "test-interface-1"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "svm_name", "svm0"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ipspace", "Default"),
				),
			},
			// Update and Read (with Home Port & Node)
			{
				Config: testAccNetworkIPInterfaceResourceConfigHomePortNode("svm0", "10.10.10.20", "e0d", "ontap_cluster_1-01", "default-data-iscsi", "Default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "name", "test-interface-1"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ip.address", "10.10.10.20"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ipspace", "Default"),
				),
			},
			// Test importing a resource (with Home Port & Node)
			{
				ResourceName:  "netapp-ontap_network_ip_interface.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "test-interface-1", "svm0", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "name", "test-interface-1"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ip.address", "10.10.10.20"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "location.home_port", "e0d"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "location.home_node", "ontap_cluster_1-01"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ipspace", "Default"),
				),
			},

			// Create and Read (with Broadcast Domain)
			{
				Config: testAccNetworkIPInterfaceResourceConfigBroadcastDomain("svm0", "10.10.20.10", "tf_test_data_svm0", "default-data-files", "Default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "name", "test-interface-2"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "svm_name", "svm0"),
					resource.TestMatchResourceAttr("netapp-ontap_network_ip_interface.example", "location.broadcast_domain.id", regexp.MustCompile(idRegexNetworkIPInterface)),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "location.home_node", "ontap_cluster_1-01"),
					resource.TestMatchResourceAttr("netapp-ontap_network_ip_interface.example", "location.home_port", regexp.MustCompile("e0a-[0-9]+")),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ipspace", "Default"),
				),
			},
			// Update and Read (with Broadcast Domain)
			{
				Config: testAccNetworkIPInterfaceResourceConfigBroadcastDomain("svm0", "10.10.20.20", "tf_test_data_svm0", "default-data-iscsi", "Default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "name", "test-interface-2"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ip.address", "10.10.20.20"),
					resource.TestMatchResourceAttr("netapp-ontap_network_ip_interface.example", "location.broadcast_domain.id", regexp.MustCompile(idRegexNetworkIPInterface)),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "location.home_node", "ontap_cluster_1-01"),
					resource.TestMatchResourceAttr("netapp-ontap_network_ip_interface.example", "location.home_port", regexp.MustCompile("e0a-[0-9]+")),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ipspace", "Default"),
				),
			},
			// Test importing a resource (with Home Port & Node)
			{
				ResourceName:  "netapp-ontap_network_ip_interface.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "test-interface-2", "svm0", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "name", "test-interface-2"),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ip.address", "10.10.20.20"),
					resource.TestMatchResourceAttr("netapp-ontap_network_ip_interface.example", "location.broadcast_domain.id", regexp.MustCompile(idRegexNetworkIPInterface)),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "location.home_node", "ontap_cluster_1-01"),
					resource.TestMatchResourceAttr("netapp-ontap_network_ip_interface.example", "location.home_port", regexp.MustCompile("e0a-[0-9]+")),
					resource.TestCheckResourceAttr("netapp-ontap_network_ip_interface.example", "ipspace", "Default"),
				),
			},
		},
	})
}

func testAccNetworkIPInterfaceProviderConfig() string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
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
	`, host, admin, password)
}

func testAccNetworkIPInterfaceResourceConfigHomePortNode(svmName, address, homePort, homeNode, servicePolicy string,ipspace string) string {
	return testAccNetworkIPInterfaceProviderConfig() + fmt.Sprintf(`
resource "netapp-ontap_network_ip_interface" "example" {
	cx_profile_name = "cluster4"
	name = "test-interface-1"
	svm_name = "%s"
	ip = {
		address = "%s"
		netmask = 18
	}
	location = {
		home_port = "%s"
		home_node = "%s"
	}
	service_policy = "%s"
	ipspace = "%s"
}
`, svmName, address, homePort, homeNode, servicePolicy, ipspace)
}

func testAccNetworkIPInterfaceResourceConfigBroadcastDomain(svmName, address, broadcastDomain, servicePolicy string,ipspace string) string {
	return testAccNetworkIPInterfaceProviderConfig() + fmt.Sprintf(`
resource "netapp-ontap_network_ip_interface" "example" {
	cx_profile_name = "cluster4"
	name = "test-interface-2"
	svm_name = "%s"
	ip = {
		address = "%s"
		netmask = 18
	}
	location = {
		broadcast_domain = {
			name = "%s"
		}
	}
	service_policy = "%s"
	ipspace = "%s"
}
`, svmName, address, broadcastDomain, servicePolicy, ipspace)
}
