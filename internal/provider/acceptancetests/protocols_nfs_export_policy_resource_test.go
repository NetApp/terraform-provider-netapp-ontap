package acceptancetests

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNFSExportPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNFSExportPolicyResourceConfig("non-existant"),
				ExpectError: regexp.MustCompile("svm non-existant not found"),
			},
			// Read testing
			{
				Config: testAccNFSExportPolicyResourceConfig("carchi-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_nfs_export_policy_resource.example", "name", "acc_test"),
					resource.TestCheckNoResourceAttr("netapp-ontap_protocols_nfs_export_policy_resource.example", "volname"),
				),
			},
		},
	})
}

func testAccNFSExportPolicyResourceConfig(svm string) string {
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

resource "netapp-ontap_protocols_nfs_export_policy_resource" "example" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "acc_test"
}`, host, admin, password, svm)
}
