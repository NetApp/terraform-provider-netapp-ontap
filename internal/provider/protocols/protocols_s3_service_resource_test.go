package protocols_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"	
)

func TestAccS3ServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccS3ServiceResourceBasicConfig("tf_acc_svm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "name", "tf_acc_svm.example.com"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "comment", "disabled"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "is_http_enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "is_https_enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "port", "80"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "secure_port", "443"),
				),
			},
			// update and read
			// update enabled, comment, and is_http_enabled
			{
				Config: testAccS3ServiceResourceUpdateConfig("tf_acc_svm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "name", "tf_acc_svm.example.com"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "comment", "enabled"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "is_http_enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "is_https_enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "port", "80"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "secure_port", "443"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_s3_service.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "tf_acc_svm", "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_service.example", "svm_name", "tf_acc_svm"),
				),
			},
		},
	})
}

func testAccS3ServiceResourceBasicConfig(svmName string) string {
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

resource "netapp-ontap_s3_service" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "tf_acc_svm.example.com"
  enabled = false
  comment = "disabled"
  is_https_enabled = false
}`, host, admin, password, svmName)
}

func testAccS3ServiceResourceUpdateConfig(svmName string) string {
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

resource "netapp-ontap_s3_service" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  name = "tf_acc_svm.example.com"
  enabled = true
  comment = "enabled"
  is_http_enabled = true
}`, host, admin, password, svmName)
}
