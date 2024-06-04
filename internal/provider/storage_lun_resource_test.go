package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageLunResouce(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test create storage lun svm not found
			{
				Config:      testAccStorageLunResourceConfig("ACC-lun", "unknownsvm", "lunTest", "linux", 1048576),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// Test create storage lun volume not found
			{
				Config:      testAccStorageLunResourceConfig("ACC-lun", "carchi-test", "unnownsvm", "linux", 1048576),
				ExpectError: regexp.MustCompile("917927"),
			},
			// Create storage lun and read
			{
				Config: testAccStorageLunResourceConfig("ACC-lun", "carchi-test", "lunTest", "linux", 1048576),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "logical_unit", "ACC-lun"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "svm_name", "carchi-test"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "volume_name", "lunTest"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "os_type", "linux"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "size", "1048576"),
				),
			},
			// Update name
			{
				Config: testAccStorageLunResourceConfig("ACC-lun2", "carchi-test", "lunTest", "linux", 1048576),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "logical_unit", "ACC-lun2"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "svm_name", "carchi-test"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "volume_name", "lunTest"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "os_type", "linux"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "size", "1048576"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_storage_lun_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "/vol/lunTest/ACC-import-lun", "lunTest", "carchi-test", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "logical_unit", "ACC-import-lun"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "os_type", "linux"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun_resource.example", "size", "1048576"),
				),
			},
		},
	})
}

func testAccStorageLunResourceConfig(logicalUnit string, svmname string, volumeName string, osType string, size int64) string {
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

resource "netapp-ontap_storage_lun_resource" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  logical_unit = "%s"
  svm_name = "%s"
  volume_name = "%s"
  os_type = "%s"
  size = "%d"
}`, host, admin, password, logicalUnit, svmname, volumeName, osType, size)
}
