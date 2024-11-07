package networking_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetworkBroadcastDomainResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Error 1377267: Specified IPspace not found
			{
				Config:      testAccNetworkBroadcastDomainResourceConfig("non-existent", "test-bd-1", 1500),
				ExpectError: regexp.MustCompile("Code:\"1377267\""),
			},
			// Error 2: "mtu" is a required field
			{
				Config:      testAccNetworkBroadcastDomainResourceConfig("Default", "test-bd-1", 0),
				ExpectError: regexp.MustCompile("Code:\"2\""),
			},
			// Create and Read
			{
				Config: testAccNetworkBroadcastDomainResourceConfig("Default", "test-bd-1", 1500),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "ipspace", "Default"),
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "name", "test-bd-1"),
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "mtu", "1500"),
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "ports.#", "0"),
				),
			},
			// Update and Read
			{
				Config: testAccNetworkBroadcastDomainResourceConfig("Default", "test-bd-2", 9000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "ipspace", "Default"),
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "name", "test-bd-2"),
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "mtu", "9000"),
				),
			},
			// Error 1377596:
			//   The "Cluster" IPspace cannot contain more than one broadcast domain.
			//   Modifying	the system IPspace "Cluster" is not supported.
			{
				Config:      testAccNetworkBroadcastDomainResourceConfig("Cluster", "test-bd-2", 9000),
				ExpectError: regexp.MustCompile("Code:\"1377596\""),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_network_broadcast_domain.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "cluster4", "Default", "test-bd-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "ipspace", "Default"),
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "name", "test-bd-2"),
					resource.TestCheckResourceAttr("netapp-ontap_network_broadcast_domain.example", "mtu", "9000"),
				),
			},
		},
	})
}

func testAccNetworkBroadcastDomainResourceConfig(ipspace, name string, mtu int64) string {
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

resource "netapp-ontap_network_broadcast_domain" "example" {
	cx_profile_name = "cluster4"
	ipspace = "%s"
	name = "%s"
  mtu = %d
}
`, host, admin, password, ipspace, name, mtu)
}
