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
const idRegex string = "[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}"

func TestAccNetworkEthernetPortResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Error 1967088: The specified broadcast domain name does not exist in the specified IPspace
			{
				Config:      testAccNetworkEthernetPortLAGResourceConfig("non-existent", "non-existent", "netapp_single-01", "mac", "\"e0b\",\"e0c\"", "singlemode"),
				ExpectError: regexp.MustCompile("Code:\"1967088\""),
			},
			// Error 1967085: The specified node name is not valid
			{
				Config:      testAccNetworkEthernetPortLAGResourceConfig("Default", "Default", "non-existent", "mac", "\"e0b\",\"e0c\"", "singlemode"),
				ExpectError: regexp.MustCompile("Code:\"1967085\""),
			},
			// Error 1967095: Invalid LAG member port name and node name combination
			{
				Config:      testAccNetworkEthernetPortLAGResourceConfig("Default", "Default", "netapp_single-01", "mac", "\"e0z\"", "singlemode"),
				ExpectError: regexp.MustCompile("Code:\"1967095\""),
			},
			// Error 1966466: VLAN ID must be a number from 1 to 4094
			{
				Config:      testAccNetworkEthernetPortVLANResourceConfig("Default", "Default", "netapp_single-01", "e0a", 9999),
				ExpectError: regexp.MustCompile("Code:\"1966466\""),
			},
			// Error 1967091: Invalid VLAN base port name
			{
				Config:      testAccNetworkEthernetPortVLANResourceConfig("Default", "Default", "netapp_single-01", "e0z", 300),
				ExpectError: regexp.MustCompile("Code:\"1967091\""),
			},
			// Create and Read LAG
			{
				Config: testAccNetworkEthernetPortLAGResourceConfig("Default", "Default", "netapp_single-01", "mac", "\"e0b\",\"e0c\"", "singlemode"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netapp-ontap_port.lag", "broadcast_domain.id", regexp.MustCompile(idRegex)),
					resource.TestCheckResourceAttr("netapp-ontap_port.lag", "enabled", "true"),
					resource.TestMatchResourceAttr("netapp-ontap_port.lag", "id", regexp.MustCompile(idRegex)),
					resource.TestMatchResourceAttr("netapp-ontap_port.lag", "node.id", regexp.MustCompile(idRegex)),
				),
			},
			// Refresh and Read LAG
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_port.lag", "lag.active_ports.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_port.lag", "name", "a0a"),
					resource.TestCheckResourceAttr("netapp-ontap_port.lag", "reachability", "not_repairable"),
					resource.TestCheckResourceAttr("netapp-ontap_port.lag", "state", "up"),
				),
			},
			// TODO: Update and Read LAG

			// Create and Read VLAN
			{
				Config: testAccNetworkEthernetPortVLANResourceConfig("tf_test", "tf_test_data_svm02", "netapp_single-01", "e0a", 300),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "broadcast_domain.id", regexp.MustCompile(idRegex)),
					resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "enabled", "true"),
					resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "id", regexp.MustCompile(idRegex)),
					resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "node.id", regexp.MustCompile(idRegex)),
				),
			},
			// Refresh and Read VLAN
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "name", "e0a-300"),
					resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "reachability", "not_repairable"),
					resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "state", "up"),
				),
			},
			// TODO: Update and Read VLAN

			// TODO: Test importing a resource
		},
	})
}

func testAccNetworkEthernetPortProviderConfig() string {
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
}`, host, admin, password)
}

func testAccNetworkEthernetPortLAGResourceConfig(bd_ipspace, bd_name, node_name, distribution_policy, member_ports, mode string) string {
	return testAccNetworkEthernetPortProviderConfig() + fmt.Sprintf(`
resource "netapp-ontap_port" "lag" {
	cx_profile_name = "cluster4"

	broadcast_domain = {
    ipspace = "%s"
    name    = "%s"
  }

  node = {
    name = "%s"
  }

  type = "lag"

  lag = {
    distribution_policy = "%s"
    member_ports = [%s]
    mode = "%s"
  }
}`, bd_ipspace, bd_name, node_name, distribution_policy, member_ports, mode)
}

func testAccNetworkEthernetPortVLANResourceConfig(bd_ipspace, bd_name, node_name, base_port string, tag int64) string {
	return testAccNetworkEthernetPortProviderConfig() + fmt.Sprintf(`
resource "netapp-ontap_port" "vlan" {
	cx_profile_name = "cluster4"

  broadcast_domain = {
    ipspace = "%s"
    name    = "%s"
  }

  node = {
    name = "%s"
  }

  type = "vlan"

  vlan = {
    base_port = "%s"
    tag       = %d
  }
}
`, bd_ipspace, bd_name, node_name, base_port, tag)
}
