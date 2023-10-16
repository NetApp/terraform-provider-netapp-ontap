package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var host string
var admin string
var password string

func TestAccStorageVolumeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
				Config: testAccStorageVolumeResourceConfig("automation", "accVolume1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_resource.example", "name", "accVolume1"),
					resource.TestCheckNoResourceAttr("netapp-ontap_storage_volume_resource.example", "volname"),
				),
			},
			{
				Config: testAccStorageVolumeResourceConfigUpdate("automation", "accVolume1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_resource.example", "name", "accVolume1"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_volume_resource.example", "nas.group_id", "10"),
					resource.TestCheckNoResourceAttr("netapp-ontap_storage_volume_resource.example", "volname"),
				),
			},
		},
	})
}

func testAccStorageVolumeResourceConfig(svm, volName string) string {
	if host == "" || admin == "" || password == "" {
		host = os.Getenv("TF_ACC_NETAPP_HOST2")
		admin = os.Getenv("TF_ACC_NETAPP_USER")
		password = os.Getenv("TF_ACC_NETAPP_PASS")
	}
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
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

resource "netapp-ontap_storage_volume_resource" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  svm_name = "%s"
  aggregates = ["aggr1"]
  space_guarantee = "none"
  snapshot_policy = "default-1weekly"
  encryption = true
  snaplock = {
    type = "non_snaplock"
  }
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
	if host == "" || admin == "" || password == "" {
		host = os.Getenv("TF_ACC_NETAPP_HOST2")
		admin = os.Getenv("TF_ACC_NETAPP_USER")
		password = os.Getenv("TF_ACC_NETAPP_PASS")
	}
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
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

resource "netapp-ontap_storage_volume_resource" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  svm_name = "%s"
  aggregates = ["aggr1"]
  space_guarantee = "none"
  snapshot_policy = "default-1weekly"
  encryption = true
  snaplock = {
    type = "non_snaplock"
  }
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
