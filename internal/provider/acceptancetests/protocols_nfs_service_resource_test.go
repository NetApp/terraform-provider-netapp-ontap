package acceptancetests

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccNfsServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccNfsServiceResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_nfs_service_resource.example", "svm_name", "carchi-test"),
				),
			},
		},
	})
}

const testAccNfsServiceResourceConfig = `
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

resource "netapp-ontap_protocols_nfs_service_resource" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "carchi-test"
  enabled = true
  protocol = {
    v3_enabled = false
    v40_enabled = true
    v40_features = {
      acl_enabled = true
    }
  }
}`
