package autosupport_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAutoSupportDataSource(t *testing.T) {
	t.Setenv("TF_ACC_NETAPP_HOST", "10.193.180.52")
   	t.Setenv("TF_ACC_NETAPP_USER", "admin")
   	t.Setenv("TF_ACC_NETAPP_PASS", "Netapp1!")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test reading AutoSupport configuration
			{
				Config: testAccAutoSupportDataSourceConfig("cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netapp-ontap_autosupport.example", "enabled"),
					resource.TestCheckResourceAttrSet("data.netapp-ontap_autosupport.example", "transport"),
					resource.TestCheckResourceAttrSet("data.netapp-ontap_autosupport.example", "from"),
					resource.TestCheckResourceAttrSet("data.netapp-ontap_autosupport.example", "contact_support"),
					resource.TestCheckResourceAttrSet("data.netapp-ontap_autosupport.example", "is_minimal"),
					resource.TestCheckResourceAttrSet("data.netapp-ontap_autosupport.example", "ondemand_enabled"),
					resource.TestCheckResourceAttrSet("data.netapp-ontap_autosupport.example", "smtp_encryption"),
				),
			},
		},
	})
}

func testAccAutoSupportDataSourceConfig(profileName string) string {
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
      name = "%s"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

data "netapp-ontap_autosupport" "example" {
	cx_profile_name = "%s"
}`, profileName, host, admin, password, profileName)
}


