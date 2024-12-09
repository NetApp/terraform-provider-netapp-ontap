package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProtocolsSanLunMapResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non existant
			{
				Config:      testAccProtocolsSanLunMapResourceBasicConfig("/vol/abc/ACC-import-lun", "abc", "abc"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// Create protocols_san_lun-maps and read
			// {
			// 	Config: testAccProtocolsSanLunMapResourceBasicConfig("/vol/terraform/ACC-import-lun", "terraform", "terraform"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_san_lun-map.example", "svm.name", "terraform"),
			// 	),
			// },
			// Import and read
			// {
			// 	ResourceName:  "netapp-ontap_san_lun-map.example",
			// 	ImportState:   true,
			// 	ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "terraform", "terraform", "/vol/terraform/terraform", "cluster4"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_san_lun-map.example", "svm.name", "terraform"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProtocolsSanLunMapResourceBasicConfig(lunName string, igroupName string, svmName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST5")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST5, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
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

resource "netapp-ontap_san_lun-map" "example" {
  cx_profile_name = "cluster4"
  svm = {
    name = "%s"
  }
  lun = {
    name = "%s"
  }
  igroup = {
    name = "%s"
  }
}`, host, admin, password, svmName, lunName, igroupName)
}
