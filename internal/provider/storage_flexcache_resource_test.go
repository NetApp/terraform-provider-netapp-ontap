package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageFlexcacheResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non existant SVM
			{
				Config:      testAccStorageFlexcacheResourceConfig("non-existant", "terraformTest4"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// test bad volume name
			{
				Config:      testAccStorageFlexcacheResourceConfig("non-existant", "name-cant-have-dashes"),
				ExpectError: regexp.MustCompile("917888"),
			},
			// Read testing
			{
				Config: testAccStorageFlexcacheResourceConfig("acc_test", "accFlexcache"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_flexcache_resource.example", "name", "accFlexcache"),
					resource.TestCheckNoResourceAttr("netapp-ontap_storage_flexcache_resource.example", "volname"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_storage_flexcache_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "accFlexcache", "acc_test", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_flexcache_resource.example", "name", "accFlexcache"),
				),
			},
		},
	})
}

func testAccStorageFlexcacheResourceConfig(svm, volName string) string {
	if host == "" || admin == "" || password == "" {
		host = os.Getenv("TF_ACC_NETAPP_HOST2")
		admin = os.Getenv("TF_ACC_NETAPP_USER")
		password = os.Getenv("TF_ACC_NETAPP_PASS2")
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

resource "netapp-ontap_storage_flexcache_resource" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  svm_name = "%s"

  origins = [
    {
      volume = {
        name = "acc_test_storage_flexcache_origin_volume"
      },
      svm = {
        name = "acc_test"
      }
    }
  ]
  size = 200
  size_unit = "mb"
  guarantee = {
    type = "none"
  }
  dr_cache = false
  global_file_locking_enabled = false
  aggregates = [
    {
      name = "acc_test"
    }
  ]
}`, host, admin, password, volName, svm)
}
