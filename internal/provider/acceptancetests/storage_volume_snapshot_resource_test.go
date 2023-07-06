package acceptancetests

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"regexp"
	"testing"
)

func TestAccStorageVolumeSnapshotResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// non-existant SVM return code 2621462. Must happen before create/read
			{
				Config:      testAccStorageVolumeSnapshotResourceConfig("non-existant"),
				ExpectError: regexp.MustCompile("Error: No svm found"),
			},
			// Read testing
			{
				Config: testAccStorageVolumeSnapshotResourceConfig("carchi-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "volume.name", "carchi_test_root"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "name", "snaptest"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "svm.name", "carchi-test"),
				),
			},
		},
	})
}

func testAccStorageVolumeSnapshotResourceConfig(svmName string) string {
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
  volume = {
    name = "carchi_test_root"
  }
  svm = {
    name = "%s"
  }
}`, host, admin, password, svmName)
}
