package storage_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
	"os"
	"regexp"
	"testing"
)

func TestAccStorageLunResouce(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
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
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "name", "ACC-lun"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "svm_name", "carchi-test"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "volume_name", "lunTest"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "os_type", "linux"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "size", "1048576"),
				),
			},
			// Update name
			{
				Config: testAccStorageLunResourceConfig("ACC-lun2", "carchi-test", "lunTest", "linux", 1048576),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "name", "ACC-lun2"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "svm_name", "carchi-test"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "volume_name", "lunTest"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "os_type", "linux"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "size", "1048576"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_storage_lun.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "ACC-import-lun", "lunTest", "carchi-test", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "name", "ACC-import-lun"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "os_type", "linux"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_lun.example", "size", "1048576"),
				),
			},
		},
	})
}

func testAccStorageLunResourceConfig(name string, svmname string, volumeName string, osType string, size int64) string {
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

resource "netapp-ontap_storage_lun" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "%s"
  svm_name = "%s"
  volume_name = "%s"
  os_type = "%s"
  size = "%d"
}`, host, admin, password, name, svmname, volumeName, osType, size)
}
