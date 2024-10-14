package security_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
)

func TestAccSecurityAccountResourceAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityAccountResourceConfigAlias("carchitest", "password"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_account_resource.security_account", "name", "carchitest"),
				),
			},
			// Test updating a resource
			{
				Config: testAccSecurityAccountResourceConfigAlias("carchitest", "password123"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_account_resource.security_account", "name", "carchitest"),
					resource.TestCheckResourceAttr("netapp-ontap_security_account_resource.security_account", "password", "password123"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_security_account_resource.security_account",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "acc_user", "cluster2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_account_resource.security_account", "name", "acc_user"),
				),
			},
		},
	})
}

func testAccSecurityAccountResourceConfigAlias(name string, accpassword string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST2")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster2"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_security_account_resource" "security_account" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  name = "%s"
  applications = [{
    application = "http"
    authentication_methods = ["password"]
  }]
  password = "%s"
}
`, host, admin, password, name, accpassword)
}
