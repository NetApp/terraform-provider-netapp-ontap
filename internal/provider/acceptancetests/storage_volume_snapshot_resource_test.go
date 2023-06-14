package acceptancetests

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccStorageVolumeSnapshotResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccStorageVolumeSnapshotResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "volume.name", "carchi_test_root"),
				),
			},
			{
				Config: testAccStorageVolumeSnapshotResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "name", "snaptest"),
				),
			},
			{
				Config: testAccStorageVolumeSnapshotResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "svm.name", "carchi-test"),
				),
			},
		},
	})
}

const testAccStorageVolumeSnapshotResourceConfig = `
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "10.193.180.108"
      username = "admin"
      password = "netapp1!"
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
    name = "carchi-test"
  }
}`
