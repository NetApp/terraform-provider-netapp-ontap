package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccS3GroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test error
			{
				Config:      testAccS3GroupResourceBasicConfig("non-existent", "tf_acc_s3_group"),
				ExpectError: regexp.MustCompile("svm non-existent not found"),
			},
			// Create and read
			{
				Config: testAccS3GroupResourceBasicConfig("tf_acc_svm", "tf_acc_s3_group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "name", "tf_acc_s3_group"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "users.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "users.0", "csahu_user1"),
				),
			},
			// update and read
			// update users and policies
			{
				Config: testAccS3GroupResourceUpdateConfig("tf_acc_svm", "tf_acc_s3_group", "created group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "name", "tf_acc_s3_group"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "comment", "created group"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "users.#", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "users.0", "csahu_user1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "users.1", "csahu_user2"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "policies.#", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "policies.0", "csahu_policy1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "policies.1", "csahu_policy2"),
				),
			},
			// update and read
			// rename and update comment
			{
				Config: testAccS3GroupResourceUpdateConfig("tf_acc_svm", "tf_acc_s3_group_renamed", "renamed group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "name", "tf_acc_s3_group_renamed"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "comment", "renamed group"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_s3_group.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_acc_s3_group_renamed", "tf_acc_svm", "cluster5"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "name", "tf_acc_s3_group_renamed"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "comment", "renamed group"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "users.#", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "users.0", "csahu_user1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "users.1", "csahu_user2"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "policies.#", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "policies.0", "csahu_policy1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_group.example", "policies.1", "csahu_policy2"),
				),
			},
		},
	})
}

func testAccS3GroupResourceBasicConfig(svmName, groupName string) string {
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
	  name = "cluster5"
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  validate_certs = false
	},
  ]
}

resource "netapp-ontap_s3_group" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  svm_name = "%s"
  name = "%s"
  users = ["csahu_user1"]
}`, host, admin, password, svmName, groupName)
}

func testAccS3GroupResourceUpdateConfig(svmName, groupName string, comment string) string {
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
      name = "cluster5"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_s3_group" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  svm_name = "%s"
  name = "%s"
  users = ["csahu_user1", "csahu_user2"]
  policies = ["csahu_policy1", "csahu_policy2"]
  comment = "%s"
}`, host, admin, password, svmName, groupName, comment)
}