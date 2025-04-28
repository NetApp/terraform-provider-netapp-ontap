package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const nnRegexProtocolsIscsiService string = "iqn\\.1992-08\\.com.netapp:sn.[[:xdigit:]]+:vs.[[:xdigit:]]"

func TestAccIscsiServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Error 2621462: The supplied SVM does not exist
			{
				Config:      testAccIscsiServiceResourceConfig("non-existant", true, "alias01"),
				ExpectError: regexp.MustCompile("Code:\"2621462\""),
			},
			// Create and read
			{
				Config: testAccIscsiServiceResourceConfig("tf_acc_svm", true, "alias01"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "target.alias", "alias01"),
					resource.TestMatchResourceAttr("netapp-ontap_iscsi_service.example", "target.name", regexp.MustCompile(nnRegexProtocolsIscsiService)),
				),
			},
			// Update and read
			{
				Config: testAccIscsiServiceResourceConfig("tf_acc_svm", true, "alias02"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "target.alias", "alias02"),
					resource.TestMatchResourceAttr("netapp-ontap_iscsi_service.example", "target.name", regexp.MustCompile(nnRegexProtocolsIscsiService)),
				),
			},
			// Update and read
			{
				Config: testAccIscsiServiceResourceConfig("tf_acc_svm", false, "alias02"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "target.alias", "alias02"),
					resource.TestMatchResourceAttr("netapp-ontap_iscsi_service.example", "target.name", regexp.MustCompile(nnRegexProtocolsIscsiService)),
				),
			},
			// Error 5374081: Cannot modify ISCSI service status-admin and other parameters simultaneously
			{
				Config:      testAccIscsiServiceResourceConfig("tf_acc_svm", true, "alias03"),
				ExpectError: regexp.MustCompile("Code:\"5374081\""),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_iscsi_service.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "tf_acc_svm", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_iscsi_service.example", "target.alias", "alias02"),
					resource.TestMatchResourceAttr("netapp-ontap_iscsi_service.example", "target.name", regexp.MustCompile(nnRegexProtocolsIscsiService)),
				),
			},
		},
	})
}

func testAccIscsiServiceResourceConfig(svm string, enabled bool, alias string) string {
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

resource "netapp-ontap_iscsi_service" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "%s"
  enabled = %t
  target = {
    alias = "%s"
  }
}`, host, admin, password, svm, enabled, alias)
}
