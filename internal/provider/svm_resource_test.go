package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"regexp"
	"testing"
)

func TestAccSvmResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSvmResourceConfig("tfsvm4", "test"),
				Check: resource.ComposeTestCheckFunc(
					// Check to see the svm name is correct,
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "name", "tfsvm4"),
					// Check to see if Ipspace is set correctly
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "ipspace", "ansibleIpspace_newname"),
					// Check that a UUID has been set (we don't know what the vaule is as it changes
					resource.TestCheckResourceAttrSet("netapp-ontap_svm_resource.example", "uuid"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "comment", "test")),
			},
			// Update a comment
			{
				Config: testAccSvmResourceConfig("tfsvm4", "carchi8py was here"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "comment", "carchi8py was here"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "name", "tfsvm4")),
			},
			// change SVM name
			{
				Config: testAccSvmResourceConfig("tfsvm3", "carchi8py was here"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "comment", "carchi8py was here"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_resource.example", "name", "tfsvm3")),
			},
			// Fail if the name already exist
			{
				Config:      testAccSvmResourceConfig("svm5", "carchi8py was here"),
				ExpectError: regexp.MustCompile("13434908"),
			},
		},
	})
}
func testAccSvmResourceConfig(svm, comment string) string {
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

resource "netapp-ontap_svm_resource" "example" {
  cx_profile_name = "cluster4"
  name = "%s"
  ipspace = "ansibleIpspace_newname"
  comment = "%s"
  snapshot_policy = "default-1weekly"
  //subtype = "dp_destination"
  language = "en_us.utf_8"
  aggregates = ["aggr2"]
  max_volumes = "200"
}`, host, admin, password, svm, comment)
}
