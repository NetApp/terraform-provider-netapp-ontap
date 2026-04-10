package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccS3UserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccS3UserResourceBasicConfig("tf_acc_svm", "tf_acc_s3_user"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "name", "tf_acc_s3_user"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "comment", "test s3 user"),
					resource.TestMatchResourceAttr("netapp-ontap_s3_user.example", "access_key", regexp.MustCompile("^[A-Z0-9]+$")),
				),
			},
			// update and read
			// update comment and regenerate keys
			{
				Config: testAccS3UserResourceUpdateRegenerateKeysConfig("tf_acc_svm", "tf_acc_s3_user"),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "name", "tf_acc_s3_user"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "comment", "acc test user updated"),
					resource.TestMatchResourceAttr("netapp-ontap_s3_user.example", "access_key", regexp.MustCompile("^[A-Z0-9]+$")),
				),
			},
			// update and read
			// delete the keys
			{
				Config: testAccS3UserResourceUpdateDeleteKeysConfig("tf_acc_svm", "tf_acc_s3_user"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "name", "tf_acc_s3_user"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "comment", "acc test user updated"),
					resource.TestMatchResourceAttr("netapp-ontap_s3_user.example", "access_key", regexp.MustCompile("^$")),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_s3_user.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_acc_s3_user", "tf_acc_svm", "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "name", "tf_acc_s3_user"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_user.example", "comment", "acc test user updated"),
				),
			},
		},
	})
}

func testAccS3UserResourceBasicConfig(svmName string, name string) string {
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

resource "netapp-ontap_s3_user" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  comment = "test s3 user"
  access_key = "76UTO311P7OWPSD0ZWYF"
  secret_key = "691CFsChP_CPG__5wnNPdG__OZP15_f_X41cTX83"
}`, host, admin, password, svmName, name)
}

func testAccS3UserResourceUpdateRegenerateKeysConfig(svmName string, name string) string {
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

resource "netapp-ontap_s3_user" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  comment = "acc test user updated"
  regenerate_keys = true
  key_time_to_live = "PT2M"
}`, host, admin, password, svmName, name)
}

func testAccS3UserResourceUpdateDeleteKeysConfig(svmName string, name string) string {
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

resource "netapp-ontap_s3_user" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "%s"
  comment = "acc test user updated"
  delete_keys = true
}`, host, admin, password, svmName, name)
}
