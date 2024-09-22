package storage_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVolumeFileResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create volumes_files and read
			{
				Config: testAccVolumeFileResourceBasicConfig("terraform", "terraform", "test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume_file.example", "path", "test"),
				),
			},
			// Update a option
			{
				Config: testAccVolumeFileResourceUpdateConfig("terraform", "terraform", "vol1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume_file.example", "path", "vol1"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_volume_file.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "snap_dest2", "terraform", ".snapshot", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume_file.example", "path", ".snapshot"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccVolumeFileResourceBasicConfig(volName string, svmName string, path string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST5")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST5, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
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

resource "netapp-ontap_volume_file" "example" {
  cx_profile_name = "cluster4"
  volume_name = "%s"
  svm_name = "%s"
  path = "%s"
  type = "directory"
  unix_permissions = "755"
}`, host, admin, password, volName, svmName, path)
}

func testAccVolumeFileResourceUpdateConfig(volName string, svmName string, path string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST5")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST5, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
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
	
	resource "netapp-ontap_volume_file" "example" {
	  cx_profile_name = "cluster4"
	  volume_name = "%s"
	  svm_name = "%s"
	  path = "%s"
	  type = "directory"
	  unix_permissions = "755"
}`, host, admin, password, volName, svmName, path)
}
