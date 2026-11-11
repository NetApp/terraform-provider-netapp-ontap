package networking_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIPspaceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccIPspaceResourceBasicConfig("tf_acc_ipspace"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ipspace.example", "name", "tf_acc_ipspace"),
				),
			},
			// rename ipspace
			{
				Config: testAccIPspaceResourceBasicConfig("tf_acc_ipspace_renamed"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ipspace.example", "name", "tf_acc_ipspace_renamed"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_network_ipspace.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "Default", "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_network_ipspace.example", "name", "Default"),
				),
			},
		},
	})
}

func testAccIPspaceResourceBasicConfig(name string) string {
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
	  name = "hw-cluster"
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  validate_certs = false
	},
  ]
}

resource "netapp-ontap_network_ipspace" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  name = "%s"
}`, host, admin, password, name)
}
