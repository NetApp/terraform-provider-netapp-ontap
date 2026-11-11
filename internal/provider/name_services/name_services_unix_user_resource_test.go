package name_services_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNameServicesUnixUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccNameServicesUnixUserResourceBasicConfig("tf_acc_svm", "tf_acc_unix_user"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "name", "tf_acc_unix_user"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "id", "99"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "primary_gid", "99"),
				),
			},
			// Update and read
			// update full_name, id, primary_gid
			{
				Config: testAccNameServicesUnixUserResourceUpdateConfig("tf_acc_svm", "tf_acc_unix_user"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "name", "tf_acc_unix_user"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "full_name", "test UNIX user"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "id", "100"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "primary_gid", "100"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_unix_user.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_acc_unix_user", "tf_acc_svm", "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "name", "tf_acc_unix_user"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "full_name", "test UNIX user"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "id", "100"),
					resource.TestCheckResourceAttr("netapp-ontap_unix_user.example", "primary_gid", "100"),
				),
			},
		},
	})
}

func testAccNameServicesUnixUserResourceBasicConfig(svmName string, name string) string {
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

resource "netapp-ontap_unix_user" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  primary_gid = 99
  id = 99
}`, host, admin, password, svmName, name)
}

func testAccNameServicesUnixUserResourceUpdateConfig(svmName string, name string) string {
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

resource "netapp-ontap_unix_user" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  full_name = "test UNIX user"
  primary_gid = 100
  id = 100
}`, host, admin, password, svmName, name)
}
