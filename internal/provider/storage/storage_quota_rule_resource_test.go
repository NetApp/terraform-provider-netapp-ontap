package storage_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageQuotaRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create storage_quota_rule and read
			{
				Config: testAccStorageQuotaRuleResourceBasicConfig("tf_acc_volume", "tf_acc_svm", 100, 80, 20971520, 5242880),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_quota_rule.example", "qtree.name", ""),
					resource.TestCheckResourceAttr("netapp-ontap_quota_rule.example", "files.hard_limit", "100"),
					resource.TestCheckResourceAttr("netapp-ontap_quota_rule.example", "files.soft_limit", "80"),
					resource.TestCheckResourceAttr("netapp-ontap_quota_rule.example", "space.hard_limit", "20971520"), // 20mb
					resource.TestCheckResourceAttr("netapp-ontap_quota_rule.example", "space.soft_limit", "5242880"),  // 5mb
				),
			},
			// Update a option
			{
				Config: testAccStorageQuotaRuleResourceBasicConfig("tf_acc_volume", "tf_acc_svm", 90, 70, 26214400, 10485760),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_quota_rule.example", "files.hard_limit", "90"),
					resource.TestCheckResourceAttr("netapp-ontap_quota_rule.example", "files.soft_limit", "70"),
					resource.TestCheckResourceAttr("netapp-ontap_quota_rule.example", "space.hard_limit", "26214400"), // 25mb
					resource.TestCheckResourceAttr("netapp-ontap_quota_rule.example", "space.soft_limit", "10485760"), // 10mb
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccStorageQuotaRuleResourceBasicConfig(volumeName string, svmName string, filesHardLimit int64, filesSoftLimit int64, spaceHardLimit int64, spaceSoftLimit int64) string {
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

resource "netapp-ontap_quota_rule" "example" {
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
	space = {
	  hard_limit = %v
	  soft_limit = %v
	}
  }`, host, admin, password, volumeName, svmName, filesHardLimit, filesSoftLimit, spaceHardLimit, spaceSoftLimit)
}
