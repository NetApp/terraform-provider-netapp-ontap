package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFcpServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Error 2621462: The supplied SVM does not exist
			{
				Config:      testAccFcpServiceResourceConfig("non-existant", true),
				ExpectError: regexp.MustCompile("Code:\"2621462\""),
			},
			// Create and read
			{
				Config: testAccFcpServiceResourceConfig("tf_acc_svm", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fcp_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_fcp_service.example", "enabled", "true"),
				),
			},
			// Update and read
			{
				Config: testAccFcpServiceResourceConfig("tf_acc_svm", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fcp_service.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_fcp_service.example", "enabled", "false"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_fcp_service.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "tf_acc_svm", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_fcp_service.example", "enabled", "false"),
				),
			},
		},
	})
}

func testAccFcpServiceResourceConfig(svm string, enabled bool) string {
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
      name           = "cluster4"
      hostname       = "%s"
      username       = "%s"
      password       = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_fcp_service" "example" {
  cx_profile_name = "cluster4"
  svm_name        = "%s"
  enabled         = %t
}`, host, admin, password, svm, enabled)
}
