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
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "volume.name", "terraformTest4"),
				),
			},
			{
				Config: testAccStorageVolumeSnapshotResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "name", "test"),
				),
			},
			{
				Config: testAccStorageVolumeSnapshotResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_snapshot_resource.example", "svm.name", "ansibleSVM"),
				),
			},
		},
	})
}

const testAccStorageVolumeSnapshotResourceConfig = `
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster2"
      hostname = "10.193.78.222"
      username = "admin"
      password = "netapp1!"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_storage_volume_snapshot_resource" "example" {
  cx_profile_name = "cluster2"
  name = "test"
  volume = {
    name = "terraformTest4"
  }
  svm = {
    name = "ansibleSVM"
  }
}`
