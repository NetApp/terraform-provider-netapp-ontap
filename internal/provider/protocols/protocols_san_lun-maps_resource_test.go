package protocols_test

import (
	"fmt"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProtocolsSanLunMapsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create protocols_san_lun-maps and read
			{
				Config: testAccProtocolsSanLunMapsResourceBasicConfig("/vol/lunTest/ACC-import-lun", "test", "carchi-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_san_lun-maps_resource.example", "svm.name", "carchi-test"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_protocols_san_lun-maps_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "carchi-test", "acc_test", "/vol/lunTest/ACC-import-lun", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_san_lun-maps_resource.example", "svm.name", "carchi-test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProtocolsSanLunMapsResourceBasicConfig(lunName string, igroupName string, svmName string) string {
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

resource "netapp-ontap_protocols_san_lun-maps_resource" "example" {
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
