package acceptancetests

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccSvmResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSvmResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					// Check to see a the Vserver name is correct,
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "name", "tfsvm4"),
					// Check to see if Ipspace is set correctly
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "ipspace", "ansibleIpspace_newname"),
					// Check that a UUID has been set (we don't know what the vaule is as it changes
					resource.TestCheckResourceAttrSet("netapp-ontap_svm_resource.example", "uuid")),
			},
		},
	})
}

const testAccSvmResourceConfig = `
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

resource "netapp-ontap_svm_resource" "example" {
  cx_profile_name = "cluster4"
  name = "tfsvm4"
  ipspace = "ansibleIpspace_newname"
  comment = "test"
  snapshot_policy = "default-1weekly"
  //subtype = "dp_destination"
  language = "en_us.utf_8"
  aggregates = ["aggr2"]
  max_volumes = "200"
}`
