package acceptancetests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNFSExportPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: TestAccNFSExportPolicyResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_nfs_export_policy_resource.example", "name", "acc_test"),
				),
			},
			// Check if a key that shouldn't be there isn't there
			{
				Config: TestAccNFSExportPolicyResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("netapp-ontap_protocols_nfs_export_policy_resource.example", "volname")),
			},
		},
	})
}

const TestAccNFSExportPolicyResourceConfig = `
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "10.193.176.186"
      username = "admin"
      password = "netapp1!"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_protocols_nfs_export_policy_resource" "example" {
	cx_profile_name = "cluster4"
	vserver = "automation"
	name = "acc_test"
}`
