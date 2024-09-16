package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCifsServicesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test error
			{
				Config:      testAccCifsServicesResourceConfigMissingVars("non-existant"),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// Create and read
			{
				Config: testAccCifsServicesResourceConfig("tftestcifs", "testSVM"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cifs_services.example", "svm_name", "testSVM"),
					resource.TestCheckResourceAttr("netapp-ontap_cifs_services.example", "name", "tftestcifs"),
				),
			},
			// update and read
			{
				Config: testAccCifsServicesResourceUpdateConfig("tftestcifs", "testSVM"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cifs_services.example", "svm_name", "testSVM"),
					resource.TestCheckResourceAttr("netapp-ontap_cifs_services.example", "name", "tftestcifs"),
					resource.TestCheckResourceAttr("netapp-ontap_cifs_services.example", "security.lm_compatibility_level", "ntlm_ntlmv2_krb"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_cifs_services.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s,%s", "TFCIFS", "tfsvm", "clustercifs", "cifstest", os.Getenv("TF_ACC_NETAPP_CIFS_ADDOMAIN_PASS")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cifs_services.example", "name", "TFCIFS"),
					resource.TestCheckResourceAttr("netapp-ontap_cifs_services.example", "svm_name", "tfsvm"),
				),
			},
		},
	})
}

func testAccCifsServicesResourceConfigMissingVars(svmName string) string {
	return fmt.Sprintf(`
	resource "netapp-ontap_cifs_services" "example1" {
		svm_name = "%s"
	}
	`, svmName)
}

func testAccCifsServicesResourceConfig(name string, svmName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS_CIFS")
	cifspassword := os.Getenv("TF_ACC_NETAPP_CIFS_ADDOMAIN_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, TF_ACC_NETAPP_PASS_CIFS and TF_ACC_NETAPP_CIFS_ADDOMAIN_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
	connection_profiles = [
		{
			name = "clustercifs"
			hostname = "%s"
			username = "%s"
			password = "%s"
			validate_certs = false
		},
	]
}
resource "netapp-ontap_cifs_services" "example" {
	cx_profile_name = "clustercifs"
	svm_name = "%s"
	name = "%s"
	ad_domain = {
		fqdn = "mytfdomain.com"
		user = "cifstest"
		password = "%s"
	}
}
`, host, admin, password, svmName, name, cifspassword)
}

func testAccCifsServicesResourceUpdateConfig(name string, svmName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS_CIFS")
	cifspassword := os.Getenv("TF_ACC_NETAPP_CIFS_ADDOMAIN_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, TF_ACC_NETAPP_PASS_CIFS and TF_ACC_NETAPP_CIFS_ADDOMAIN_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
	connection_profiles = [
		{
			name = "clustercifs"
			hostname = "%s"
			username = "%s"
			password = "%s"
			validate_certs = false
		},
	]
}
resource "netapp-ontap_cifs_services" "example" {
	cx_profile_name = "clustercifs"
	svm_name = "%s"
	name = "%s"
	ad_domain = {
		fqdn = "mytfdomain.com"
		user = "cifstest"
		password = "%s"
	}
	security = {
		lm_compatibility_level = "ntlm_ntlmv2_krb"
	}
}
`, host, admin, password, svmName, name, cifspassword)
}
