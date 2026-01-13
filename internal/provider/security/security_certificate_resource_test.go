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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_certificate.example", "name", "acc_test_ca_cert2"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_security_certificate.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "acc_test_ca_cert2", "acc_test_ca_cert", "server", "cluster1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_certificate.example", "name", "acc_test_ca_cert2"),
				),
			},
			// Delete testing automatically occurs in TestCase

			// Update security certificate and read
			{
				Config: testAccSecurityCertificateResourceCertificateConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_certificate.example", "name", "acc_test_ca_cert2"),
					resource.TestCheckResourceAttr("netapp-ontap_security_certificate.example", "intermediate_certificates.#", "3"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_security_certificate.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "acc_test_ca_cert2", "acc_test_ca_cert", "server", "cluster1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_certificate.example", "name", "acc_test_ca_cert2"),
				),
			},
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
  name            = "acc_test_ca_cert2"
  common_name     = "acc_test_ca_cert"
  type            = "server"
  svm_name        = "tf_acc_svm"
  intermediate_certificates = [
    "-----BEGIN CERTIFICATE-----\n INTERMEDIATE CERTIFICATE \n-----END CERTIFICATE-----",
    "-----BEGIN CERTIFICATE-----\n INTERMEDIATE CERTIFICATE \n-----END CERTIFICATE-----",
    "-----BEGIN CERTIFICATE-----\n ROOT CERTIFICATE \n-----END CERTIFICATE-----"
]
}`, host, admin, password)
}
