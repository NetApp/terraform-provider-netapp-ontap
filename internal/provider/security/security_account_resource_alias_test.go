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
				ImportStateId: fmt.Sprintf("%s,%s,%s", "vsadmin", "tf_acc_svm", "cluster2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_account_resource.security_account", "name", "vsadmin"),
				),
			},
		},
	})
}

func testAccSecurityAccountResourceConfigAlias(name string, accpassword string) string {
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
