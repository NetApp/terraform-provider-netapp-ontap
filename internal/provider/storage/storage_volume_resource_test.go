package storage_test

import (
	"fmt"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageVolumeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non existant SVM
			{
				Config:      testAccStorageVolumeResourceConfig("non-existant", "terraformTest4"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// test bad volume name
			{
				Config:      testAccStorageVolumeResourceConfig("non-existant", "name-cant-have-dashes"),
				ExpectError: regexp.MustCompile("917888"),
			},
			// Read testing
			{
				Config: testAccStorageVolumeResourceConfig("acc_test", "accVolume1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume.example", "name", "accVolume1"),
					resource.TestCheckNoResourceAttr("netapp-ontap_storage_volume.example", "volname"),
				),
			},
			{
				Config: testAccStorageVolumeResourceConfigUpdate("automation", "accVolume1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume.example", "name", "accVolume1"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume.example", "nas.group_id", "10"),
					resource.TestCheckNoResourceAttr("netapp-ontap_storage_volume.example", "volname"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_storage_volume.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "acc_test_root", "acc_test", "cluster5"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume.example", "name", "automation"),
				),
			},
		},
	})
}

func testAccStorageVolumeResourceConfig(svm, volName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST2")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")

	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster5"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_storage_volume" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  svm_name = "%s"
  aggregates = [
	{name = "acc_test"}
]
  space_guarantee = "none"
  snapshot_policy = "default-1weekly"
  space = {
	size = 30
	size_unit = "mb"
	percent_snapshot_space = 10
    logical_space = {
      enforcement = true
      reporting = true
    }
  }
  tiering = {
  	policy_name = "all"
  }
  nas = {
    export_policy_name = "test"
    group_id = 1
    user_id = 2
    unix_permissions = "100"
    security_style = "mixed"
	junction_path = "/testacc"
  }
}`, host, admin, password, volName, svm)
}

func testAccStorageVolumeResourceConfigUpdate(svm, volName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST2")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")

	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster5"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_storage_volume" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  svm_name = "%s"
  aggregates = [
	{name = "acc_test"}
]
  space_guarantee = "none"
  snapshot_policy = "default-1weekly"
  space = {
	size = 30
	size_unit = "mb"
	percent_snapshot_space = 20
    logical_space = {
      enforcement = true
      reporting = true
    }
  }
  tiering = {
  	policy_name = "all"
  }
  nas = {
    export_policy_name = "test"
    group_id = 10
    user_id = 20
    unix_permissions = "755"
    security_style = "mixed"
	junction_path = "/testacc"
  }
}`, host, admin, password, volName, svm)
}
