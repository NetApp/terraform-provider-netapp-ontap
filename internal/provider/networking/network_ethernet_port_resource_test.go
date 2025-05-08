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
// const idRegexNetworkEthernetPort string = "[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}"

func TestAccNetworkEthernetPortResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Error 1967088: The specified broadcast domain name does not exist in the specified IPspace
			{
				Config:      testAccNetworkEthernetPortLAGResourceConfig("non-existent", "non-existent", true, "bsuhas-vsim1", "mac", "\"e0b\",\"e0c\"", "singlemode"),
				ExpectError: regexp.MustCompile("Code:\"1967088\""),
			},
			// Error 1967085: The specified node name is not valid
			{
				Config:      testAccNetworkEthernetPortLAGResourceConfig("Default", "Default", true, "non-existent", "mac", "\"e0b\",\"e0c\"", "singlemode"),
				ExpectError: regexp.MustCompile("Code:\"1967085\""),
			},
			// Error 1967095: Invalid LAG member port name and node name combination
			{
				Config:      testAccNetworkEthernetPortLAGResourceConfig("Default", "Default", true, "bsuhas-vsim1", "mac", "\"e0z\"", "singlemode"),
				ExpectError: regexp.MustCompile("Code:\"1967095\""),
			},
			// Error 1966466: VLAN ID must be a number from 1 to 4094
			// {
			// 	Config:      testAccNetworkEthernetPortVLANResourceConfig("Default", "Default", true, "bsuhas-vsim1", "e0a", 9999),
			// 	ExpectError: regexp.MustCompile("Code:\"1966466\""),
			// },
			// // Error 1967091: Invalid VLAN base port name
			// {
			// 	Config:      testAccNetworkEthernetPortVLANResourceConfig("Default", "Default", true, "bsuhas-vsim1", "e0z", 300),
			// 	ExpectError: regexp.MustCompile("Code:\"1967091\""),
			// },

			// Create and Read LAG
			// {
			// 	Config: testAccNetworkEthernetPortLAGResourceConfig("Default", "Default", true, "bsuhas-vsim1", "mac", "\"e0b\",\"e0c\"", "singlemode"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.lag", "broadcast_domain.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.lag", "id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "lag.active_ports.#", "1"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "name", "a0a"),
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.lag", "node.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "state", "up"),
			// 	),
			// },
			// // Update (disable) and Read LAG
			// {
			// 	Config: testAccNetworkEthernetPortLAGResourceConfig("tf_test", "tf_test_data_svm02", false, "bsuhas-vsim1", "mac", "\"e0b\"", "singlemode"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.lag", "broadcast_domain.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "lag.active_ports.#", "0"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "state", "down"),
			// 	),
			// },
			// // Update (enable) and Read LAG
			// {
			// 	Config: testAccNetworkEthernetPortLAGResourceConfig("tf_test", "tf_test_data_svm02", true, "bsuhas-vsim1", "mac", "\"e0b\"", "singlemode"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "lag.active_ports.#", "1"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "lag.active_ports.0", "e0b"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "state", "up"),
			// 	),
			// },
			// // Test importing LAG
			// {
			// 	ResourceName:  "netapp-ontap_port.lag",
			// 	ImportState:   true,
			// 	ImportStateId: fmt.Sprintf("%s,%s", "cluster4", "a0a"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.lag", "broadcast_domain.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "broadcast_domain.name", "tf_test"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "broadcast_domain.name", "tf_test_data_svm02"),
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.lag", "id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "lag.active_ports.0", "e0b"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "lag.distribution_policy", "mac"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "lag.mode", "singlemode"),
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.lag", "node.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "node.name", "bsuhas-vsim1"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.lag", "state", "up"),
			// 	),
			// },

			// // Create and Read VLAN
			// {
			// 	Config: testAccNetworkEthernetPortVLANResourceConfig("tf_test", "tf_test_data_svm02", true, "bsuhas-vsim1", "e0a", 300),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "broadcast_domain.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "name", "e0a-300"),
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "node.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "reachability", "not_repairable"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "state", "up"),
			// 	),
			// },
			// // Update (disable) and Read VLAN
			// {
			// 	Config: testAccNetworkEthernetPortVLANResourceConfig("Default", "Default", false, "bsuhas-vsim1", "e0a", 300),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "broadcast_domain.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "state", "down"),
			// 	),
			// },
			// // Update (disable) and Read VLAN
			// {
			// 	Config: testAccNetworkEthernetPortVLANResourceConfig("Default", "Default", true, "bsuhas-vsim1", "e0a", 300),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "state", "up"),
			// 	),
			// },
			// // Test importing VLAN
			// {
			// 	ResourceName:  "netapp-ontap_port.vlan",
			// 	ImportState:   true,
			// 	ImportStateId: fmt.Sprintf("%s,%s", "cluster4", "e0a-300"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "broadcast_domain.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "broadcast_domain.name", "Default"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "broadcast_domain.name", "Default"),
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestMatchResourceAttr("netapp-ontap_port.vlan", "node.id", regexp.MustCompile(idRegexNetworkEthernetPort)),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "node.name", "bsuhas-vsim1"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "state", "up"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "vlan.base_port", "e0a"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_port.vlan", "vlan.tag", "300"),
			// 	),
			// },
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

func testAccNetworkEthernetPortLAGResourceConfig(bd_ipspace, bd_name string, enabled bool, node_name, distribution_policy, member_ports, mode string) string {
	return testAccNetworkEthernetPortProviderConfig() + fmt.Sprintf(`
resource "netapp-ontap_port" "lag" {
	cx_profile_name = "cluster4"

	broadcast_domain = {
    ipspace = "%s"
    name    = "%s"
  }

	enabled = %t

  node = {
    name = "%s"
  }

  type = "lag"
  lag = {
    distribution_policy = "%s"
    member_ports = [%s]
    mode = "%s"
  }
}`, bd_ipspace, bd_name, enabled, node_name, distribution_policy, member_ports, mode)
}

// func testAccNetworkEthernetPortVLANResourceConfig(bd_ipspace, bd_name string, enabled bool, node_name, base_port string, tag int64) string {
// 	return testAccNetworkEthernetPortProviderConfig() + fmt.Sprintf(`
// resource "netapp-ontap_port" "vlan" {
// 	cx_profile_name = "cluster4"

//   broadcast_domain = {
//     ipspace = "%s"
//     name    = "%s"
//   }

// 	enabled = %t

//   node = {
//     name = "%s"
//   }

//   type = "vlan"
//   vlan = {
//     base_port = "%s"
//     tag       = %d
//   }
// }
// `, bd_ipspace, bd_name, enabled, node_name, base_port, tag)
// }
