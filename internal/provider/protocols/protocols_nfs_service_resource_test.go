package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNfsServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test error
			{
				Config:      testAccNfsServiceResourceConfig("non-existent", "false"),
				ExpectError: regexp.MustCompile("svm non-existent not found"),
			},
			// Create and read
			{
				Config: testAccNfsServiceResourceConfig("tf_acc_svm", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v3_enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v40_enabled", "true"),
				),
			},
			// update and read
			{
				Config: testAccNfsServiceResourceConfig("tf_acc_svm", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v3_enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v40_enabled", "true"),
				),
			},
			// update and read
			{
				Config: testAccNfsServiceResourceConfigV41Disabled("tf_acc_svm", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v3_enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v41_enabled", "false"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_nfs_service.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "tf_acc_svm", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v3_enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_nfs_service.example", "protocol.v40_enabled", "true"),
				),
			},
		},
	})
}

func testAccNfsServiceResourceConfig(svnName, enableV3 string) string {
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

resource "netapp-ontap_nfs_service" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "%s"
  enabled = true
  protocol = {
    v3_enabled = "%s"
    v40_enabled = true
    v40_features = {
      acl_enabled = true
    }
  }
}`, host, admin, password, svnName, enableV3)
}

func testAccNfsServiceResourceConfigV41Disabled(svnName, disableV41 string) string {
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

resource "netapp-ontap_nfs_service" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "%s"
  enabled = true
  protocol = {
    v3_enabled = true
    v40_enabled = true
	v41_enabled = "%s"
    v40_features = {
      acl_enabled = true
    }
  }
}`, host, admin, password, svnName, disableV41)
}
