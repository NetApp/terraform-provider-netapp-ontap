package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageSnapshotPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test create storage snapshot policy error
			{
				Config:      testAccStorageSnapshotPolicyResourceConfig("non-existant", "unknowsvm", "wrong case", false),
				ExpectError: regexp.MustCompile("error creating storage_snapshot_policy"),
			},
			// Create storage snapshot policy and read
			{
				Config: testAccStorageSnapshotPolicyResourceConfig("tf-sn-policy", "carchi-test", "create a test snapshot policy", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_snapshot_policy_resource.example", "name", "tf-sn-policy"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_snapshot_policy_resource.example", "comment", "create a test snapshot policy"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_snapshot_policy_resource.example", "enabled", "true"),
				),
			},
			// Update storage snapshot policy on comment and read
			{
				Config: testAccStorageSnapshotPolicyResourceConfig("tf-sn-policy", "carchi-test", "Update the existing snapshot policy", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_snapshot_policy_resource.example", "name", "tf-sn-policy"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_snapshot_policy_resource.example", "comment", "Update the existing snapshot policy"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_snapshot_policy_resource.example", "enabled", "true"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_storage_snapshot_policy_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tfimportpolicy", "carchi-test", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_nfs_export_policy_resource.example", "name", "tfimportpolicy"),
				),
			},
		},
	})
}

func testAccStorageSnapshotPolicyResourceConfig(name string, svmname string, comment string, enabled bool) string {
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

resource "netapp-ontap_storage_snapshot_policy_resource" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "%s"
  svm_name = "%s"
  comment = "%s"
  enabled = "%t"
  copies = [
  {
	count = 3
	schedule = {
	  name = "daily"
	}
  },
  ]
}`, host, admin, password, name, svmname, comment, enabled)
}
