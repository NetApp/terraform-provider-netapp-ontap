package storage_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageVolumeCloneResource(t *testing.T) {
	cloneName := fmt.Sprintf("tf_acc_volume_clone_%d", time.Now().UnixNano())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccStorageVolumeCloneResourceBasicConfig("tf_acc_svm", cloneName, cloneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume_clone.example", "name", cloneName),
					resource.TestCheckResourceAttr("netapp-ontap_volume_clone.example", "svm_name", "tf_acc_svm"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_volume_clone.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", cloneName, "tf_acc_svm", "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_volume_clone.example", "name", cloneName),
					resource.TestCheckResourceAttr("netapp-ontap_volume_clone.example", "svm_name", "tf_acc_svm"),
				),
			},
		},
	})
}

func testAccStorageVolumeCloneResourceBasicConfig(svmName string, name string, path string) string {
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
	  name = "hw-cluster"
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  validate_certs = false
	},
  ]
}

resource "netapp-ontap_volume_clone" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  clone = {
	parent_volume = "tf_acc_volume"
  }
  nas = {
	junction_path = "/%s"
	group_id = 1
	user_id = 1
  }
}`, host, admin, password, svmName, name, path)
}
