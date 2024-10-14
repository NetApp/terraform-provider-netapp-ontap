package protocols_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProtocolsSanIgroupResourceAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create protocols_san_igroup and read
			{
				Config: testAccProtocolsSanIgroupResourceBasicConfigAlias("acc_test2", "carchi-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_san_igroup_resource.example", "name", "acc_test2"),
				),
			},
			// Update options and read
			{
				Config: testAccProtocolsSanIgroupResourceUpdateConfigAlias("acc_test2", "carchi-test", "windows", "test_acc"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_san_igroup_resource.example", "os_type", "windows"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_san_igroup_resource.example", "name", "acc_test2"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_protocols_san_igroup_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "acc_test", "carchi-test", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_san_igroup_resource.example", "name", "acc_test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProtocolsSanIgroupResourceBasicConfigAlias(name string, svmName string) string {
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

resource "netapp-ontap_protocols_san_igroup_resource" "example" {
  cx_profile_name = "cluster4"
  name = "%s"
  svm = {
    name = "%s"
  }
  os_type = "linux"
  comment = "test"
}`, host, admin, password, name, svmName)
}

func testAccProtocolsSanIgroupResourceUpdateConfigAlias(name string, svmName string, osType string, comment string) string {
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

resource "netapp-ontap_protocols_san_igroup_resource" "example" {
  cx_profile_name = "cluster4"
  name = "%s"
  svm = {
    name = "%s"
  }
  os_type = "%s"
  comment = "%s"
}`, host, admin, password, name, svmName, osType, comment)
}
