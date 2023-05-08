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
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "ipspace", "test"),
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
      name = "cluster2"
      hostname = "10.193.78.222"
      username = "admin"
      password = "netapp1!"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_svm_resource" "example" {
  cx_profile_name = "cluster2"
  name = "tfsvm4"
  ipspace = "test"
  comment = "test"
  snapshot_policy = "default-1weekly"
  //subtype = "dp_destination"
  language = "en_us.utf_8"
  aggregates = ["aggr1", "test"]
  max_volumes = "200"
}`
