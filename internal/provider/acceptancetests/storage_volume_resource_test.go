package acceptancetests

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccStorageVolumeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccStorageVolumeResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_resource.example", "name", "terraformTest4"),
				),
			},
			// Check if a key that shouldn't be there isn't there
			{
				Config: testAccStorageVolumeResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("netapp-ontap_storage_volume_resource.example", "volname")),
			},
		},
	})
}

const testAccStorageVolumeResourceConfig = `
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

resource "netapp-ontap_storage_volume_resource" "example" {
  cx_profile_name = "cluster2"
  name = "terraformTest4"
  vserver = "ansibleSVM"
  aggregates = ["aggr1"]
  size = 20
  size_unit = "mb"
}`
