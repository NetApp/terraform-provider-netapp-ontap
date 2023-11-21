package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageVolumeSnapshotResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// non-existant SVM return code 2621462. Must happen before create/read
			{
				Config:      testAccStorageVolumeSnapshotResourceConfig("non-existant", "my comment"),
				ExpectError: regexp.MustCompile("Error: No svm found"),
			},
			// Create and read testing
			{
				Config: testAccStorageVolumeSnapshotResourceConfig("carchi-test", "my comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "volume_name", "carchi_test_root"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "name", "snaptest"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "svm_name", "carchi-test"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "comment", "my comment"),
				),
			},
			// Update and read testing
			{
				Config: testAccStorageVolumeSnapshotResourceConfig("carchi-test", "new comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "volume_name", "carchi_test_root"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "name", "snaptest"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "svm_name", "carchi-test"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "comment", "new comment"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_storage_volume_snapshot_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "snaptest", "carchi_test_root", "carchi-test", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "name", "snaptest"),
				),
			},
		},
	})
}

func testAccStorageVolumeSnapshotResourceConfig(svmName string, comment string) string {
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

resource "netapp-ontap_storage_volume_snapshot_resource" "example" {
  cx_profile_name = "cluster4"
  name = "snaptest"
  volume_name = "carchi_test_root"
  svm_name = "%s"
  comment = "%s"
}`, host, admin, password, svmName, comment)
}
