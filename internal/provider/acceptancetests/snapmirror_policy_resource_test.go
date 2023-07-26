package acceptancetests

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"regexp"
	"testing"
)

func TestAccSnapmirrorPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSnapmirrorPolicyResourceConfig("non-existant"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			{
				Config: testAccSnapmirrorPolicyResourceConfig("ansibleSVM"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "name", "carchitestme4"),
				),
			},
		},
	})
}

func testAccSnapmirrorPolicyResourceConfig(svm string) string {
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

resource "netapp-ontap_snapmirror_policy_resource" "example" {
  cx_profile_name = "cluster4"
  name = "carchitestme4"
  svm_name = "%s"
  identity_preservation = "full"
  comment = "comment1"
  type = "async"
}`, host, admin, password, svm)
}
