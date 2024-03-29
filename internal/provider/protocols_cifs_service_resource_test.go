package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCifsServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test error
			{
				Config:      testAccCifsServiceResourceConfigMissingVars("non-existant"),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// Create and read
			{
				Config: testAccCifsServiceResourceConfig("tftestcifs", "testSVM"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "svm_name", "testSVM"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "name", "tftestcifs"),
				),
			},
			// update and read
			{
				Config: testAccCifsServiceResourceUpdateConfig("tftestcifs", "testSVM"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "svm_name", "testSVM"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "name", "tftestcifs"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "security.lm_compatibility_level", "ntlm_ntlmv2_krb"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_protocols_cifs_service_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s,%s", "TFCIFS", "tfsvm", "clustercifs", "administrator", "Ab0xB@wks!"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "name", "TFCIFS"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "svm_name", "tfsvm"),
				),
			},
		},
	})
}

func testAccCifsServiceResourceConfigMissingVars(svmName string) string {
	return fmt.Sprintf(`
	resource "netapp-ontap_protocols_cifs_service_resource" "example1" {
		svm_name = "%s"
	}
	`, svmName)
}

func testAccCifsServiceResourceConfig(name string, svmName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS_CIFS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS_CIFS must be set for acceptance tests")
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
resource "netapp-ontap_protocols_cifs_service_resource" "example" {
	cx_profile_name = "clustercifs"
	svm_name = "%s"
	name = "%s"
	ad_domain = {
		fqdn = "mytfdomain.com"
		user = "administrator"
		password = "Ab0xB@wks!"
	}
}
`, host, admin, password, svmName, name)
}

func testAccCifsServiceResourceUpdateConfig(name string, svmName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS_CIFS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS_CIFS must be set for acceptance tests")
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
resource "netapp-ontap_protocols_cifs_service_resource" "example" {
	cx_profile_name = "clustercifs"
	svm_name = "%s"
	name = "%s"
	ad_domain = {
		fqdn = "mytfdomain.com"
		user = "administrator"
		password = "Ab0xB@wks!"
	}
	security = {
		lm_compatibility_level = "ntlm_ntlmv2_krb"
	}
}
`, host, admin, password, svmName, name)
}
