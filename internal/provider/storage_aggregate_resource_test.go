package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageAggregateResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccStorageAggregateResourceConfig("non-existant"),
				ExpectError: regexp.MustCompile("is an invalid value"),
			},
			{
				Config: testAccStorageAggregateResourceConfig("swenjun-vsim1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_aggregate_resource.example", "name", "acc_test_aggr"),
					resource.TestCheckNoResourceAttr("netapp-ontap_storage_aggregate_resource.example", "vol"),
				),
			},
		},
	})
}

func testAccStorageAggregateResourceConfig(node string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST2")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST2, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
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

resource "netapp-ontap_storage_aggregate_resource" "example" {
	cx_profile_name = "cluster4"
	node = "%s"
	name = "acc_test_aggr"
	disk_count = 5
	disk_size = 1
	disk_size_unit= "gb"
	is_mirrored = false
	raid_type = "raid4"
	snaplock_type = "compliance"
	encryption = true
}`, host, admin, password, node)
}
