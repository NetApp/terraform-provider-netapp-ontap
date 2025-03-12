package storage_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageVolumeEfficiencyPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create volume_efficiency_policy and read
			{
				Config: testAccStorageVolumeEfficiencyPolicyResourceBasicConfig("testacc", "tf_acc_svm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume_efficiency_policy.example", "name", "testacc"),
				),
			},
			// Update a option
			{
				Config: testAccStorageVolumeEfficiencyPolicyResourceUpdateConfig("testacc", "tf_acc_svm", "test_comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume_efficiency_policy.example", "comment", "test_comment"),
					resource.TestCheckResourceAttr("netapp-ontap_volume_efficiency_policy.example", "name", "testacc"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_volume_efficiency_policy.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "default", "tf_acc_svm", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume_efficiency_policy.example", "name", "default"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccStorageVolumeEfficiencyPolicyResourceBasicConfig(name string, svmName string) string {
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

resource "netapp-ontap_volume_efficiency_policy" "example" {
  cx_profile_name = "cluster4"
  name = "%s"
  svm = {
	name = "%s"
  }
}`, host, admin, password, name, svmName)
}

func testAccStorageVolumeEfficiencyPolicyResourceUpdateConfig(name string, svmName string, comment string) string {
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

resource "netapp-ontap_volume_efficiency_policy" "example" {
  cx_profile_name = "cluster4"
  name = "%s"
  svm = {
	name = "%s"
  }
  comment = "%s"
}`, host, admin, password, name, svmName, comment)
}
