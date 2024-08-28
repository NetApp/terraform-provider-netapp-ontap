package storage_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageQuotaRulesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create storage_quota_rules and read
			{
				Config: testAccStorageQuotaRulesResourceBasicConfig("lunTest", "carchi-test", 100, 80),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_quota_rules.example", "qtree.name", ""),
				),
			},
			// Update a option
			{
				Config: testAccStorageQuotaRulesResourceBasicConfig("lunTest", "carchi-test", 100, 70),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_quota_rules.example", "files.hard_limit", "100"),
					resource.TestCheckResourceAttr("netapp-ontap_quota_rules.example", "files.soft_limit", "70"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_quota_rules.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s,%s", "lunTest", "carchi-test", "tree", "testacc", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_quota_rules.example", "name", "name2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccStorageQuotaRulesResourceBasicConfig(volumeName string, svmName string, hardLimit int64, softLimit int64) string {
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

resource "netapp-ontap_quota_rules" "example" {
	cx_profile_name = "cluster4"
	volume = {
	  name = "%s"
	  }
	svm = {
	  name = "%s"
	  }
	type = "tree"
	qtree = {
	  name = ""
	  }
	files = {
	  hard_limit = %v
	  soft_limit = %v
	  }
  }`, host, admin, password, volumeName, svmName, hardLimit, softLimit)
}
