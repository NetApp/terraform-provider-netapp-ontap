package protocols_test

import (
	"fmt"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNfsServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test error
			{
				Config:      testAccNfsServiceResourceConfig("non-existant", "false"),
				ExpectError: regexp.MustCompile("svm non-existant not found"),
			},
			// Create and read
			{
				Config: testAccNfsServiceResourceConfig("carchi-test", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "svm_name", "carchi-test"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v3_enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v40_enabled", "true"),
				),
			},
			// update and read
			{
				Config: testAccNfsServiceResourceConfig("carchi-test", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "svm_name", "carchi-test"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v3_enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v40_enabled", "true"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_nfs_service.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "carchi-test", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v3_enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v40_enabled", "true"),
				),
			},
		},
	})
}

func testAccNfsServiceResourceConfig(svnName, enableV3 string) string {
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

resource "netapp-ontap_nfs_service" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "%s"
  enabled = true
  protocol = {
    v3_enabled = "%s"
    v40_enabled = true
    v40_features = {
      acl_enabled = true
    }
  }
}`, host, admin, password, svnName, enableV3)
}
