package security_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSecurityCertificateResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create security certificate and read
			{
				Config: testAccSecurityCertificateResourceCertificateConfig(),
				Check:  resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_certificate.example", "name", "tfsvm_ca_cert1"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_security_certificate.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "tfsvm_ca_cert1", "cluster1"),
				Check:         resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_certificate.example", "name", "tfsvm_ca_cert1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSecurityCertificateResourceCertificateConfig() string {
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
      name = "cluster1"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_security_certificate" "example" {
  cx_profile_name = "cluster1"
  name            = "tfsvm_ca_cert1"
  common_name     = "tfsvm_ca_cert"
  type            = "root_ca"
  svm_name        = "tfsvm"
}`, host, admin, password)
}
